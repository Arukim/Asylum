package asylum

import (
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"bufio"
	"math/rand"
	"strconv"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var guestBook []string

const (
	stateOffline = iota
	stateServerOverflow = iota
	stateConnected = iota
	stateInGame = iota
)

type Money struct{
	Coin int
	Potion int
}

// Types related to cards
type Treasure struct{
	Vl Money
}

type Action struct{
	Ap int
	Bp int
	Vl Money
	Specials []int
}

type Victory struct{
	Wp int
}

type Card struct{
	Tr Treasure
	Ac Action
	Vc Victory
	Cost Money
	Known bool
	Name string
}

type Player struct{
	Hand []Card
	Discard []Card
	Deck []Card
}


// The whole game struct
type Table struct{
	piles [][]Card
	players []Player
}

type Bot struct{
	Name string
	GamesCount int
	LastGameDuration string
	LastGameTurnCount int
	AvgTurnSpeed string
	AvgGameDuration string
	State string
	state int
	buyer Buyer
	conn net.Conn
	remoteAddr string
	currTurnCount int
	lastGameDuration time.Duration
	sumGameDuration time.Duration
	sumGameTurns int
}

//Communication with server
type LoginPacket struct{
	Name string `json:"name"`
}

type ServerInfoPacket struct{
	Version string `json:"version"`
}

type ServerOptions struct{
	Type string `json:"type"`
	Targets []string `json:"targets"`
}

type ServerTurnPacket struct{
	Piles [] ServerPile `json:"piles"`
	TurnStats ServerTurnStats `json:"turnStats"`
	Options [] ServerOptions `json:"options"`
	Players [] ServerPlayer `json:"players"`
	LastEvent ServerOptions `json:"lastEvent"`
	LastEventPlayer string `json:"lastEventPlayer"`
}

type ServerPlayer struct{
	Hand [] string `json:"hand"`
	Name string `json:"name"`
	TopDiscard string `json:"topDiscardCardName"`
	DeckSize int `json:"nofCardsInDeck"`
}

type ServerPile struct{
	TopCardName string `json:"topCardName"`
	PrototypeName string `json:"prototypeName"`
	Cards int `json:"nofCards"`
}
type ServerTurnStats struct{
	ActivePlayerName string `json:"activePlayerName"`
	TurnNumber int `json:"turnNumber"`
	Play [] string `json:"play"`
	Actions int `json:"nofActions"`
	Buys int `"nofBuys"`
	Coins int `"nofCoins"`
	Potions int `"nofPotions"`
}

type ServerFinalCard struct{
	Card string `json:"card"`
	Count int `json:"count"`
}
type ServerFinalGameStats struct{
	Name string `json:"name"`
	Points int `json:"nofVictoryPoints"`
	State string `json:"state"`
	Cards ServerFinalCard `json:"cards"`
}

type ClientTurnPacket struct{
	OptionNumber int `json:"optionNumber"`
}

// Players Logic
type Buyer interface {
	GetName() string
	Buy(* Bot,* ServerTurnPacket) ClientTurnPacket
	
}

type ChaoticBuyer struct{
	name string
}
type GreedyBuyer struct{
	name string
}

func (buyer ChaoticBuyer) GetName() string{
	return "Chaotic"
}

func (_ ChaoticBuyer) Buy (bot* Bot, turnInfo* ServerTurnPacket) ClientTurnPacket{
	var clientPacket ClientTurnPacket
	turnCount := len(turnInfo.Options)
	clientPacket.OptionNumber = rand.Intn(turnCount)
	return clientPacket
}


func (buyer GreedyBuyer) GetName() string{
	return "Greedy"
}
func (_ GreedyBuyer) Buy(bot* Bot, turnInfo* ServerTurnPacket) ClientTurnPacket{
	var clientPacket ClientTurnPacket
	myTurn := 0
TurnChoosed:
	for i,option := range turnInfo.Options {
		switch option.Type {
		case "PLAY_ALL_TREASURES":
			myTurn = i
			break TurnChoosed
		case "BUY":
			switch option.Targets[0] {
			case "Province":
				myTurn = i
				break TurnChoosed
			case "Gold":
				myTurn = i
				break TurnChoosed
			case "Silver":
				myTurn = i
				break TurnChoosed
			}
		}
	}
	clientPacket.OptionNumber = myTurn
	return clientPacket
}
// Auxilary
func indexOf(arr []string, target string) int{
	for i, elem := range(arr){
		if(elem == target){
			return i
		}
	}
	return -1
}
// Init	
var table Table
var CardsPool = []Card{}
var stateList map[int]string
func init(){
	db, err := sql.Open("sqlite3", "db/log.sqlite3")
	if err != nil {
		log.Println("DB open err", err)
	}else{
		defer db.Close()
	}
	err = db.Ping()
	if err != nil {
		log.Println("Can't ping DB")
	}

	stateList = map[int]string{
			stateOffline:"ServerOffline",
			stateConnected:"Connected",
			stateServerOverflow:"ServerOverflow",
			stateInGame:"InGame",
	}

	jsonBlob, err := ioutil.ReadFile("db/cards.json")
	if err != nil {
		log.Fatal(err)
	}
	if err:= json.Unmarshal(jsonBlob, &CardsPool); err != nil {
		log.Fatal(err)
	}

	jsonBlob, err = ioutil.ReadFile("db/guestbook.json")
	if err != nil {
		log.Fatal(err)
	}
	if err:= json.Unmarshal(jsonBlob, &guestBook); err != nil {
		log.Fatal(jsonBlob,err)
	}
}

func statistics(bot* Bot){
	if bot.currTurnCount > 0 {
		bot.GamesCount++
		bot.LastGameTurnCount = bot.currTurnCount
		bot.sumGameDuration += bot.lastGameDuration
		bot.sumGameTurns += bot.LastGameTurnCount
		bot.AvgTurnSpeed = fmt.Sprintf("%.03f s", bot.sumGameDuration.Seconds() / float64(bot.sumGameTurns))
		bot.AvgGameDuration = fmt.Sprintf("%.03f s", bot.sumGameDuration.Seconds() / float64(bot.GamesCount))
		bot.currTurnCount = 0
	}
}

// Bot state handlers
func hBotOffline(bot* Bot){
	for bot.state == stateOffline {
		_conn, err := net.Dial("tcp", bot.remoteAddr)
		if err != nil {
			time.Sleep(1 * time.Second)
			//log.Println("Can't resolve server address")
		}else{
			bot.conn = _conn
			bot.state = stateConnected;
			statistics(bot)
		}
	}
}

func hBotConnected(bot* Bot){
	for bot.state == stateConnected {
		str, err := bufio.NewReader(bot.conn).ReadString('\n')
		if err != nil {
			log.Println("Can't read %v", err)
			bot.state = stateServerOverflow
		}else{
			var serverInfo ServerInfoPacket
			err := json.Unmarshal([]byte(str), &serverInfo)
			if err != nil{
				log.Println("error:", err)
				log.Println(str)
			}else{
				var packet LoginPacket
				packet.Name = bot.Name
				wr, _ := json.Marshal(packet)
				_, _ = bot.conn.Write([]byte(wr))
				bot.state = stateInGame
			}
		}
	}
}

func hBotInGame(bot* Bot){
	bot.currTurnCount = 0
	gameStarted := time.Now()
	bufRead := bufio.NewReader(bot.conn)
	for bot.state == stateInGame {
		str, err := bufRead.ReadString('\n')
		if err != nil {
			bot.state = stateOffline
			now := time.Now()
			bot.lastGameDuration = now.Sub(gameStarted)
			bot.LastGameDuration = fmt.Sprintf("%.03f s", bot.lastGameDuration.Seconds())

		}else{
			var turnPacket ServerTurnPacket
			err = json.Unmarshal([]byte(str), &turnPacket)
			if err != nil {
				log.Println("error:", err)
				log.Println(str)
			}else{
				if len(turnPacket.Options) != 0 {
					bot.currTurnCount++
					clientTurn := bot.buyer.Buy(bot, &turnPacket)
					wr, _ := json.Marshal(clientTurn)
					_, _ = bot.conn.Write([]byte(wr))
				}
			}
			var finalStats []ServerFinalGameStats
			err = json.Unmarshal([]byte(str), &finalStats)
			if err != nil {
			}else{
				if len(finalStats) > 0 {
					log.Println("Final stats:", finalStats)
				}
			}
		}
	}
}


func updateState(bot * Bot){
	bot.State = stateList[bot.state]
}

func generateName(bot * Bot){
	name := guestBook[rand.Intn(len(guestBook))]
	bot.Name = bot.buyer.GetName() + " " + name + " "  +strconv.Itoa(rand.Int() % 1000)
}

func (bot Bot) Born(remoteAddr string, uplink chan Bot){
	if rand.Intn(10) > 7{
		bot.buyer = new(ChaoticBuyer)
	}else{
		bot.buyer = new(GreedyBuyer)
	}
	bot.remoteAddr = remoteAddr
	generateName(&bot)
	for {
		updateState(&bot)
		uplink <- bot
		switch bot.state{
		case stateOffline:
			hBotOffline(&bot)
		case stateConnected:
			hBotConnected(&bot)
		case stateInGame:
			hBotInGame(&bot)
		case stateServerOverflow:
			time.Sleep(1 * time.Second)
			bot.state = stateOffline
		default:
			panic("Unknown state!");
		}			
	}
}