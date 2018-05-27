package main

import (
	"io/ioutil"
	"os"

	"time"

	"bytes"
	"net/http"

	"encoding/json"

	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// Global Configuration
type config struct {
	Logger           *zap.Logger
	Port             string                 `yaml:"port"`
	Debug            string                 `yaml:"debug"`
	EncKey           string                 `yaml:"encKey"`
	Remote           string                 `yaml:"remote"`
	ExpHours         int64                  `yaml:"expHours"`
	GetTokenRoute    string                 `yaml:"getTokenRoute"`
	CheckTokenRoute  string                 `yaml:"checkTokenRoute"`
	RequestTokenData map[string]interface{} `yaml:"requestTokenData"`
}

func main() {

	logger, _ := zap.NewDevelopment()

	configFile := getEnv("CFG_FILE", "example_cfg.yml")
	cfg, err := loadConfiguration(configFile)
	if err != nil {
		logger.Panic(err.Error())
	}

	// set the logger
	cfg.Logger = logger

	gin.SetMode(gin.ReleaseMode)

	// set gin to debug mode if env DEBUG=true
	if cfg.Debug == "true" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// check encryption key length
	if len(cfg.EncKey) < 32 {
		logger.Panic("ENCKEY is too short.")
	}

	// set zap to production mode if we are not debugging
	// and no longer in a setup phase
	if cfg.Debug != "true" {
		logger, _ = zap.NewProduction()
		cfg.Logger = logger
	}

	// use zap logger middleware
	r.Use(ginzap.Ginzap(cfg.Logger, time.RFC3339, true))

	// add encryption key to the context
	r.Use(func(c *gin.Context) {
		c.Set("Cfg", cfg)
		c.Next()
	})

	// Route get token request
	r.POST(cfg.GetTokenRoute, tokenRouteHandler)

	// Route check token request
	r.GET(cfg.CheckTokenRoute, checkTokenRouteHandler)

	r.Run(":" + cfg.Port)
}

func checkTokenRouteHandler(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config)

	decoded, err := request.ParseFromRequest(c.Request, request.OAuth2Extractor, func(token *jwt_lib.Token) (interface{}, error) {
		b := []byte(cfg.EncKey)
		return b, nil
	})

	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, decoded)
}

// RouteHandler handles the http route for inbound data
func tokenRouteHandler(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config)

	remoteAuthToken := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	remoteAuthToken.Claims = jwt_lib.MapClaims{
		"data": cfg.RequestTokenData,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}

	// generate a signed token for the remote request
	remoteTokenString, err := remoteAuthToken.SignedString([]byte(cfg.EncKey))
	if err != nil {
		cfg.Logger.Error("Could not generate remoteAuthToken: " + err.Error())
		c.JSON(500, gin.H{"message": "Could not generate remoteAuthToken", "error": err.Error()})
		return
	}

	// make remote request passing rawData from the post to us
	rawData, err := c.GetRawData()
	if err != nil {
		cfg.Logger.Error("Unable to read post data: " + err.Error())
		c.JSON(500, gin.H{"message": "Unable to read post data", "error": err.Error()})
		return
	}

	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("POST", cfg.Remote, bytes.NewBuffer(rawData))
	req.Header.Set("Authorization", "Bearer "+remoteTokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := netClient.Do(req)
	if err != nil {
		cfg.Logger.Error("Remote post failure: " + err.Error())
		c.JSON(500, gin.H{"message": "Remote post failure.", "error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// check the remote call status code
	if resp.StatusCode != http.StatusOK {
		cfg.Logger.Error("Remote post status failure got: " + string(resp.StatusCode))
		c.JSON(500, gin.H{"message": "Remote post status failure.", "status": resp.StatusCode})
		return
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cfg.Logger.Error("Remote post failed to parse body: " + err.Error())
		c.JSON(500, gin.H{"message": "Remote post failed to parse body.", "error": err.Error()})
		return
	}

	// convert json returned from remote into a map
	retMap := make(map[string]interface{})

	err = json.Unmarshal(bodyBytes, &retMap)
	if err != nil {
		cfg.Logger.Error("Can not unmarshal returned json: " + err.Error())
		c.JSON(500, gin.H{"message": "Can not unmarshal returned json.", "error": err.Error()})
		return
	}

	token := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	token.Claims = jwt_lib.MapClaims{
		"data": retMap,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}

	tokenString, err := token.SignedString([]byte(cfg.EncKey))
	if err != nil {
		cfg.Logger.Error("Can not generate return token: " + err.Error())
		c.JSON(500, gin.H{"message": "Can not generate return token", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"token": tokenString})
}

// loadConfiguration
func loadConfiguration(filename string) (cfg config, err error) {
	cfg = config{}
	ymlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal([]byte(ymlData), &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

// getEnv gets an environment variable or sets a default if
// one does not exist.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}
