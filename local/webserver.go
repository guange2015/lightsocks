package local

import (
	"fmt"
	"github.com/guange2015/lightsocks/cmd"
	"html"
	"log"
	"net/http"
)

var _config *cmd.Config

func configHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Password, %s\n", _config.Password)
	fmt.Fprintf(w, "ListenAddr, %s\n", _config.ListenAddr)
	fmt.Fprintf(w, "RemoteAddr, %s\n", _config.RemoteAddr)
}

func StartWebServer(config *cmd.Config) {

	_config = config

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/config", configHandler)

	log.Println("start web server, listen on :1098")
	log.Fatal(http.ListenAndServe(":1098", nil))
}
