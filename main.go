package main

import (
	"time"
	"net/http"
	"html/template"
	"asylum/asylum"
	"math/rand"
	"strconv"
	)

var templates = template.Must(template.ParseFiles("tmpl/mainPage.html", "tmpl/cardsPool.html"))
var names = [...]string{"Jonn", "Piter", "Lob", "Eddie"}
var botList = []*asylum.Bot{}
type Page struct {
	Title string
	BotList []*asylum.Bot
	CardsPool *[]asylum.Card
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

func cardPool(w http.ResponseWriter, r *http.Request){
	p := &Page{Title: "CardList"}
	p.CardsPool = &asylum.CardsPool
	renderTemplate(w, "cardsPool", p)
}

func botAdd(w http.ResponseWriter, r *http.Request){
	defer http.Redirect(w, r, "../", http.StatusFound)
	bot := new(asylum.Bot)
	bot.Name = names[rand.Intn(len(names))] + strconv.Itoa(rand.Int())
	go bot.Born("hello", 1*time.Millisecond)
	botList = append(botList, bot)
}

func init(){
	rand.Seed(time.Now().UTC().UnixNano())
}

func spawnPool(){
	for i:=0; i< 4; i++{
	bot := new(asylum.Bot)
	bot.Name = names[rand.Intn(len(names))] + strconv.Itoa(rand.Int())
	go bot.Born("hello", 1*time.Millisecond)
	botList = append(botList, bot)
	}
}

func main(){
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/addBot/", botAdd)
	http.HandleFunc("/cardsPool/", cardPool)
	go spawnPool()
	http.ListenAndServe(":8080",nil)

}