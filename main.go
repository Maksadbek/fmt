package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/burntsushi/toml"
)

type Config struct {
	ServerAddr     string
	MaxInputLength int64
}

type App struct {
	templates *template.Template
	config    Config
}

func (a *App) init(configFile, tmplDir string) {
	if _, err := toml.DecodeFile(configFile, &a.config); err != nil {
		panic(err)
	}

	var err error
	if a.templates, err = template.ParseFiles(
		filepath.Join(tmplDir, "index.tmpl"),
		filepath.Join(tmplDir, "output.tmpl"),
	); err != nil {
		panic(err)
	}
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	a.templates.ExecuteTemplate(w, "index", nil)
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

	c := struct {
		Result template.JS
	}{}

	c.Result = template.JS(buf.String())
	a.templates.ExecuteTemplate(w, "output", c)
}

var (
	configFile = flag.String("config", "./etc/config.toml", "config file path")
	tmplDir    = flag.String("tmplDir", "./templates", "templates directory")
)

func main() {
	flag.Parse()

	app := new(App)
	app.init(*configFile, *tmplDir)

	http.HandleFunc("/", app.indexHandler)
	http.HandleFunc("/format/json", app.formatJSON)

	println("starting web server:", app.config.ServerAddr)
	panic(http.ListenAndServe(app.config.ServerAddr, nil))
}
