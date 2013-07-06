package main

import (
	"time"
	"net/http"
	"html/template"
	"asylum/asylum"
	"math/rand"
	"strconv"
	"flag"
	"log"
	"strings"
	)

var templates = template.Must(template.ParseFiles("tmpl/mainPage.html", "tmpl/cardsPool.html"))
var names = [...]string{"Jonn", "Piter", "Lob", "Eddie"}
var server Server

type Server struct{
	BotList [] *asylum.Bot
	Downlink chan asylum.Bot
	ReadRequest chan int
}

type Page struct {
	Title string
	BotList []*asylum.Bot
	CardsPool *[]asylum.Card
}

func botAdd(){
	bot := new(asylum.Bot)
	name := names[rand.Intn(len(names))] + strconv.Itoa(rand.Int())
	go bot.Born(serverAddr, name, server.Downlink)
}


func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
	err := templates.ExecuteTemplate(w, tmpl + ".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, ".") {
		shttp.ServeHTTP(w, r)
	}else{
		p := &Page{Title: "Welcome"}
		p.BotList = server.BotList
		renderTemplate(w, "mainPage", p)
	}
}

func cardPool(w http.ResponseWriter, r *http.Request){
	p := &Page{Title: "CardList"}
	p.CardsPool = &asylum.CardsPool
	renderTemplate(w, "cardsPool", p)
}

func hBotAdd(w http.ResponseWriter, r *http.Request){
	defer http.Redirect(w, r, "../", http.StatusFound)
	botAdd()
}

var serverAddr string
var membersCount int
func init(){
	rand.Seed(time.Now().UTC().UnixNano())
	flag.StringVar(&serverAddr, "server", "192.168.1.2:6666", "enter remote server address")
	flag.IntVar(&membersCount, "num", 4, "starting members of asylum count")
}

func spawningPool(){
	for i:=0; i< membersCount; i++{
		botAdd()
	}
}

func Collect(){
	for {
		botUpdate := <- server.Downlink
	//	log.Println("Hui")
		updated := false
		for i, bot := range server.BotList{
			if bot.Name == botUpdate.Name {
				server.BotList[i] = &botUpdate
				updated = true
			}
		}
		if !updated {
			server.BotList = append(server.BotList, &botUpdate)
		}
	}
}

var shttp = http.NewServeMux()

func main(){

	flag.Parse()
	log.Println("Remote server addr is ", serverAddr)
	log.Println("Members: ", membersCount)

	server.BotList = []*asylum.Bot{}
	server.Downlink = make(chan asylum.Bot, membersCount*10)
	go Collect()
	shttp.Handle("/", http.FileServer(http.Dir("./static/")))
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/addBot/", hBotAdd)
	http.HandleFunc("/cardsPool/", cardPool)

	go spawningPool()
	http.ListenAndServe(":8080",nil)

}