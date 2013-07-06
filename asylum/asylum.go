package asylum

import (
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"bufio"
//	"math/rand"
)

const (
	stateOffline = iota
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
	LastGameLength string
	State int
	conn net.Conn
	remoteAddr string
}

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
	Options []ServerOptions `json:"options"`
}

type ClientTurnPacket struct{
	OptionNumber int `json:"optionNumber"`
}

type Buyer interface {
	Buy(bot* Bot, turnPacket * ServerTurnPacket) ClientTurnPacket
}

type GreedyBuyer struct{
}
 
func (_ GreedyBuyer) Buy(bot* Bot, turnPacket * ServerTurnPacket) ClientTurnPacket{
	var clientPacket ClientTurnPacket
	myTurn := 0
TurnChoosed:
	for i,option := range turnPacket.Options {
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
			case "Duchy":
				myTurn = i
				break TurnChoosed
			case "Silver":
				myTurn = i
				break TurnChoosed
			case "Copper":
				myTurn = i
				break TurnChoosed
			case "Estate":
				myTurn = i
				break TurnChoosed
			}
		}
	}
	clientPacket.OptionNumber = myTurn
	return clientPacket
}
	
var table Table
var CardsPool = []Card{}
func init(){
	jsonBlob, err := ioutil.ReadFile("db/cards.json")
	if err != nil {
		log.Fatal(err)
		return
	}
	if err:= json.Unmarshal(jsonBlob, &CardsPool); err != nil {
		log.Fatal(err)
		return
	}
//	fmt.Printf("%+v", CardsPool)
}

func hBotOffline(bot* Bot){
	for bot.State == stateOffline {
		_conn, err := net.Dial("tcp", bot.remoteAddr)
		if err != nil {
			//log.Println("Can't resolve server address")
		}else{
			bot.conn = _conn
			bot.State = stateConnected;
		}
	}
}

func hBotConnected(bot* Bot){
	for bot.State == stateConnected {
		str, err := bufio.NewReader(bot.conn).ReadString('\n')
		if err != nil {
			log.Println("Can't read %v", err)
			bot.State = stateOffline
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
				bot.State = stateInGame
			}
		}
	}
}

func makeTurn(bot* Bot, turnPacket * ServerTurnPacket) ClientTurnPacket{
	var clientTurn ClientTurnPacket
/*
	turnCount := len(turnPacket.Options)
	choosed := -1
	for i,option := range turnPacket.Options{
		if option.Type == "PLAY_ALL_TREASURES" {
			choosed = i
		}
		if option.Type == "BUY" && choosed == -1 {
			if option.Target == "Estate" {
				choosed = i
			}
		}
	}
	if(choosed == -1){
		clientTurn.OptionNumber = rand.Intn(turnCount)
	}else{
		clientTurn.OptionNumber = choosed
	}*/
	return clientTurn
}

func hBotInGame(bot* Bot){
	bot.GamesCount++
	var buyer Buyer = new(GreedyBuyer)
	gameStarted := time.Now()
	bufRead := bufio.NewReader(bot.conn)
	for bot.State == stateInGame {
		str, err := bufRead.ReadString('\n')
		if err != nil {
			bot.State = stateOffline
			now := time.Now()
			bot.LastGameLength = fmt.Sprintf("%.03f s",now.Sub(gameStarted).Seconds())
		}else{
			var turnPacket ServerTurnPacket
			err := json.Unmarshal([]byte(str), &turnPacket)
			if err != nil {
				log.Println("error:", err)
				log.Println(str)
			}else{
				if len(turnPacket.Options) != 0 {
					clientTurn := buyer.Buy(bot, &turnPacket)
					wr, _ := json.Marshal(clientTurn)
					_, _ = bot.conn.Write([]byte(wr))
				}
			}
		}
	}
}

func (bot Bot) Born(remoteAddr string, name string, uplink chan Bot){
	bot.Name = name
	bot.remoteAddr = remoteAddr
	for {
		uplink <- bot
		switch bot.State{
		case stateOffline:
			hBotOffline(&bot)
		case stateConnected:
			hBotConnected(&bot)
		case stateInGame:
			hBotInGame(&bot)
		default:
			panic("Unknown state!");
		}			
	}
}