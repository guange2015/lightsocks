linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o lightsocks-linux
	scp ./lightsocks-linux ss:~/

osx:
	go build -o lightsocks-osx

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o lightsocks-win.exe



