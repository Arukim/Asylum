package asylum

import (
	"fmt"
	"time"
)

type Bot struct{
	Name string
}

func (bot Bot) Born(message string, delay time.Duration){
	for {
		time.Sleep(delay)
		fmt.Printf("%v is alive\r\n",bot.Name)
	}
}