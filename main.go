package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

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

func (a *App) init() {
	if _, err := toml.DecodeFile("./etc/config.toml", &a.config); err != nil {
		panic(err)
	}

	var err error
	if a.templates, err = template.ParseFiles(
		"./templates/index.tmpl",
		"./templates/output.tmpl",
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

	/*
		buf, err := ioutil.ReadAll(io.LimitReader(r.Body, a.config.MaxInputLength))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/

	println(string(r.Form.Get("text")))
	buf := &bytes.Buffer{}
	err := json.Indent(buf, []byte(r.Form.Get("text")), "", "   ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(buf.String())

	c := struct {
		Result template.JS
	}{}

	c.Result = template.JS(buf.String())
	a.templates.ExecuteTemplate(w, "output", c)
}

func main() {
	app := new(App)

	app.init()

	http.HandleFunc("/", app.indexHandler)
	http.HandleFunc("/format/json", app.formatJSON)

	panic(http.ListenAndServe(app.config.ServerAddr, nil))
}
