package asylum

import (
	"fmt"
	"time"
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
	vl Money
}

type Action struct{
	ap int
	bp int
	vl Money
	specials []int
}

type Victory struct{
	wp int
}

type Card struct{
	tr Treasure
	ac Action
	vc Victory
	cost Money
	known bool
}

type Player struct{
	hand []Card
	discard []Card
	deck []Card
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

func (bot Bot) Born(message string, delay time.Duration){
	for {
		switch bot.State{
		case stateOffline:
			time.Sleep(delay)
			fmt.Printf("%v is alive\r\n",bot.Name)
		default:
			time.Sleep(delay)
		}			
	}
}