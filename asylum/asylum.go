package asylum

import (
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
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

var _conn net.Conn
func (bot Bot) Born(message string, delay time.Duration){
	buf := make([]byte, 2048);
	service := "192.168.1.2:6666"
	for {
		switch bot.State{
		case stateOffline:
			time.Sleep(delay)
			conn, err := net.Dial("tcp", service)
			if err != nil {
				log.Println("Can't resolve server address")
			}else{
				_conn = conn
				bot.State = stateConnected;
			}
			fmt.Printf("%v is alive\r\n",bot.Name)
		case stateConnected:
			var packet LoginPacket
			packet.Name = bot.Name
			wr, _ := json.Marshal(packet)
			_, _ = _conn.Write([]byte(wr))
			log.Println("Writed %v", string(wr))
			for bot.State == stateConnected {
				len, err := _conn.Read(buf)
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
						}
					}
				}
			}
		default:
			time.Sleep(delay)
		}			
	}
}