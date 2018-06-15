# Ellation Reactions Service
---
Reactions Service is responsible for managing websocket connections for the livestream feature.

### Key features

* Connect to the pub/sub service via websocket
* Get stats about current number of connections in the websocket.
* Publish messages (emojis) to the websocket
* Receive broadcasted data about these messages (emojis)

### Prerequisites

* Go >= 1.10
* [dep](https://github.com/golang/dep)

### Installation

Clone this repo into ``$GOPATH/src/github.com/crunchyroll/cx-reactions``

Run `dep ensure` to install all dependencies

### Build & Run

1. Prepare `config.yaml`. A sample can be found [here](config/config.sample.yaml)
2. Build project using `go build .`
3. Run project using `go run ./cx-reactions`

### Run tests

Unit tests:
```bash
go test ./...
```

Coverage:
```bash
./coverage.sh
```

Code check:
```bash
./code-check.sh
```

### cURLs

To get stats for current connected clients to the websocket:
```curl
curl -X GET  \
    'http://localhost:8080/stats'
```

Establish websocket connection:
```curl
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Host: localhost:8080" \
     --header "Origin: http://localhost:8080" \
     --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     --header "Sec-WebSocket-Version: 13" \
     http://localhost:8080/ws
```
