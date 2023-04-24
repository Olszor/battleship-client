package main

import (
	"battleship-client/app"
	"battleship-client/http"
	"fmt"
	"log"
	"time"
)

func main() {
	client := http.NewClient("https://go-pjatk-server.fly.dev/api", time.Second*30)

	err := client.InitGame()
	if err != nil {
		log.Fatal(err)
	}

	status, err := client.Status()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(status.GameStatus)

	board, err := client.Board()
	if err != nil {
		log.Fatal(err)
	}
	app.Render(board)
}
