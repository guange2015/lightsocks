@echo off
set CGO_ENABLED=0
set GOARCH=amd64
set GOOS=darwin
go build -o lightsocks-osx
pause