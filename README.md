![txtoken data transmission](mast.jpg)
[![txtoken Release](https://img.shields.io/github/release/txn2/txtoken.svg)](https://github.com/txn2/txtoken/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/txn2/txtoken)](https://goreportcard.com/report/github.com/txn2/txtoken)
[![Docker Container Image Size](https://shields.beevelop.com/docker/image/image-size/txn2/txtoken/latest.svg)](https://hub.docker.com/r/txn2/irsync/)
[![Docker Container Layers](https://shields.beevelop.com/docker/image/layers/txn2/txtoken/latest.svg)](https://hub.docker.com/r/txn2/irsync/)
[![Docker Container Pulls](https://img.shields.io/docker/pulls/txn2/txtoken.svg)](https://hub.docker.com/r/txn2/txtoken/)

# txToken

[txToken] creates an HS256 JWT token from JSON retrieved from a remote endpoint resulting from a proxied json request body.

[txToken] exposes an http POST endpoint accepting JSON data to be sent to a remote server along with a preset bearer token. Validation of this token on the remote side is optional and only needed if the remote wishes to authenticate the call using a shared key.

1. Post JSON data to txToken.
2. txToken re-posts the JSON to remote endpoint.
3. txToken creates a JWT token with JSON data returned from remote.
4. txToken returns a JWT token

Systems that share an encryption key with txToken can validate the token and ensure the authenticity of it's data.

## Getting Started

Checkout the article [JWT Microservice, Tokenize Remote APIs] for a detailed functional overview and guide.


[JWT Microservice, Tokenize Remote APIs]: https://mk.imti.co/jwt-microservice/
[txtoken]: https://github.com/txn2/txtoken
