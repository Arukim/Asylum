package asylum

import (
	"fmt"
	"time"
)


func Born(message string, delay time.Duration){
	go func() {
		for {
			time.Sleep(delay)
			fmt.Println(message)
		}
	}()
}