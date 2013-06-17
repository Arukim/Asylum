package main

import (
	"time"
	"net/http"
	"html/template"
	"asylum/asylum"
	)

var templates = template.Must(template.ParseFiles("tmpl/mainPage.html"))

type Page struct {
	Title string
	Body []byte
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
	err := templates.ExecuteTemplate(w, tmpl + ".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{Title: "Welcome"}
	renderTemplate(w, "mainPage", p)
}

func main(){
	go asylum.Born("Hello", 1*time.Second)
	http.HandleFunc("/", mainPage)
	http.ListenAndServe(":8080",nil)
}