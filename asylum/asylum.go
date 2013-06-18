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