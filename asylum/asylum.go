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


func (bot Bot) Born(message string, delay time.Duration){
	buf := make([]byte, 2048)
//	bufPos := 0
	service := "192.168.1.2:6666"
	var conn net.Conn
	for {
		switch bot.State{
		case stateOffline:
			time.Sleep(delay)
			_conn, err := net.Dial("tcp", service)
			if err != nil {
				log.Println("Can't resolve server address")
			}else{
				conn = _conn
				bot.State = stateConnected;
			}
			fmt.Printf("%v is alive\r\n",bot.Name)
		case stateConnected:
			var packet LoginPacket
			packet.Name = bot.Name
			wr, _ := json.Marshal(packet)
			_, _ = conn.Write([]byte(wr))
			log.Println("Writed %v", string(wr))
			for bot.State == stateConnected {
				len, err := conn.Read(buf)
				if err != nil {
					log.Println("Can't read %v", err)
					bot.State = stateOffline
				}else{
					if len > 0 {
						log.Println("Readed ",string( buf))
						var serverInfo ServerInfoPacket
						err := json.Unmarshal(buf[0:len], &serverInfo)
						if err != nil{
							log.Println("error:", err)
						}else{
							log.Println(serverInfo)
							bot.State = stateInGame
						}
					}
				}
			}
		case stateInGame:
			errCount := 0
			for bot.State == stateInGame {
				str, err := bufio.NewReader(conn).ReadString('\n')
			//	len, err := conn.Read(buf)
				if err != nil {
					errCount++
					time.Sleep(delay)
					if errCount == 2 {
						bot.State = stateOffline
						log.Println("Server is dead ", err)
					}
				}else{
					errCount = 0
					//if len > 0 {
						//log.Println("Readed ",string( buf))
						var turnPacket ServerTurnPacket
					err := json.Unmarshal([]byte(str), &turnPacket)
						if err != nil {
							log.Println("error:", err)
						}else{
						turnCount := len(turnPacket.Options)
							log.Println(turnPacket)
							var clientTurn ClientTurnPacket
//leng := (turnPacket.Options)
						clientTurn.OptionNumber = rand.Intn(turnCount)
							wr, _ := json.Marshal(clientTurn)
							_, _ = conn.Write([]byte(wr))
							log.Println("Writed %v", string(wr))
						}
				/*	}else{
						errCount++
						time.Sleep(delay)
						if errCount == 5 {
							bot.State = stateOffline
							log.Println("Server is silent")
						}
					}*/
				}
			}
		default:
			time.Sleep(delay)
		}			
	}
}