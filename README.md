# txToken

txToken creates an HS256 JWT token from JSON retrieved from a remote endpoint resulting from a proxied json request body.

txToken exposes an http POST endpoint accepting JSON data to be sent to a remote server along with a preset bearer token. Validation of this token on the remote side is optional and only needed if the remote wishes to authenticate the call using a shared key.

1. Post JSON data to txToken.
2. txToken re-posts the JSON to remote endpoint.
3. txToken creates a JWT token with JSON data returned from remote.
4. txToken returns a JWT token

Systems that share an encryption key with txToken can validate the token and ensure the authenticity of it's data.


