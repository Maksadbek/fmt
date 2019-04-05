package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"

	"github.com/burntsushi/toml"
)

var (
	configFile = flag.String("config", "./etc/config.toml", "config file path")
)

type Config struct {
	ServerAddr     string
	MaxInputLength int64
}

type App struct {
	config Config
}

func (a *App) init(configFile string) {
	if _, err := toml.DecodeFile(configFile, &a.config); err != nil {
		panic(err)
	}
}

var indexPage = []byte(`<!DOCTYPE html>
<html>
  <head>
    <style>
      textarea {
        font-size: 16px;
        font-family: Courier New, Arial;
        width: 100%;
        height: 100vw;
      }
    </style>
  </head>
  
  <body>
    <form action="/format/json" method="post">
      <div><input type="submit" /></div>
      </br>
      <textarea name="text"></textarea>
    </form>
  </body>
</html>`)

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(indexPage)
}

func (a *App) formatJSON(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	defer r.Body.Close()

	buf := &bytes.Buffer{}
	err := json.Indent(buf, []byte(r.Form.Get("text")), "", "   ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
}

func main() {
	flag.Parse()

	app := new(App)
	app.init(*configFile)

	http.HandleFunc("/", app.indexHandler)
	http.HandleFunc("/format/json", app.formatJSON)

	println("starting web server:", app.config.ServerAddr)
	panic(http.ListenAndServe(app.config.ServerAddr, nil))
}
