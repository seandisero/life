package main

import (
	"fmt"

	"github.com/seandisero/life/internal/life"
)

func main() {
	life_game, err := life.Init_life_game()
	if err != nil {
		fmt.Println("could not create life:", err)
		return
	}
	life_game.Run()
}
