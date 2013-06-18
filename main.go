package main

import (
	"time"
	"net/http"
	"html/template"
	"asylum/asylum"
	"math/rand"
	)

var templates = template.Must(template.ParseFiles("tmpl/mainPage.html"))
var names = [...]string{"Jonn", "Piter", "Lob", "Eddie"}
var botList = []*asylum.Bot{}
type Page struct {
	Title string
	BotList []*asylum.Bot
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
	err := templates.ExecuteTemplate(w, tmpl + ".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{Title: "Welcome"}
	p.BotList = botList
	renderTemplate(w, "mainPage", p)
}

func botAdd(w http.ResponseWriter, r *http.Request){
	defer http.Redirect(w, r, "../", http.StatusFound)
	bot := new(asylum.Bot)
	bot.Name = names[rand.Intn(len(names))]
	go bot.Born("hello", 1*time.Second)
	botList = append(botList, bot)
}

func init(){
	rand.Seed(time.Now().UTC().UnixNano())
}

func main(){
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/addBot/", botAdd)
	http.ListenAndServe(":8080",nil)
}