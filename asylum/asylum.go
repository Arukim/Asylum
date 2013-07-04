package asylum

import (
//	"fmt"
//	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"bufio"
	"math/rand"
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
	State int
}

type LoginPacket struct{
	Name string `json:"name"`
}

type ServerInfoPacket struct{
	Version string `json:"version"`
}

type ServerOptions struct{
	Type string `json:"type"`
	Target string `json:"target"`
}

type ServerTurnPacket struct{
	Options []ServerOptions `json:"options"`
}

type ClientTurnPacket struct{
	OptionNumber int `json:"optionNumber"`
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


func (bot Bot) Born(remoteAddr string, name string, uplink chan Bot){
	bot.Name = name
	uplink <- bot
	var conn net.Conn
	for {
		switch bot.State{
		case stateOffline:
			uplink <- bot
			for bot.State == stateOffline {
				_conn, err := net.Dial("tcp", remoteAddr)
				if err != nil {
					//log.Println("Can't resolve server address")
				}else{
					conn = _conn
					bot.State = stateConnected;
				}
			}
		case stateConnected:
			uplink <- bot
			for bot.State == stateConnected {
				str, err := bufio.NewReader(conn).ReadString('\n')
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
						_, _ = conn.Write([]byte(wr))
						bot.State = stateInGame
					}
				}
			}
		case stateInGame:
			bot.GamesCount++
			uplink <- bot
			for bot.State == stateInGame {
				str, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
				//	log.Println("error:", err)
					bot.State = stateOffline
				//	time.Sleep(10*time.Millisecond)
					
				}else{
					var turnPacket ServerTurnPacket
					err := json.Unmarshal([]byte(str), &turnPacket)
					if err != nil {
						log.Println("error:", err)
						log.Println(str)
					}else{
						turnCount := len(turnPacket.Options)
						var clientTurn ClientTurnPacket
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
						}
							wr, _ := json.Marshal(clientTurn)
							_, _ = conn.Write([]byte(wr))
						}
				}
			}
		default:
		}			
	}
}