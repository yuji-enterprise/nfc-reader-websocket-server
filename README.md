# nfc-reader-websocket-server

### This app will listen card reader event via pcscd, then send card uid to websocket client.

Build golang bin follow below for mac amd64 (only test on macos)
```
# 64-bit
$ CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o ./nfc-reader-websocket-server ./main.go

# 32-bit
$ CGO_ENABLED=1 GOOS=darwin GOARCH=386 go build -o ./nfc-reader-websocket-server ./main.go
```
