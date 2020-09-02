@echo off
set CGO_ENABLED=0
set GOARCH=amd64
set GOOS=windows
go build -o lightsocks-win.exe
pause