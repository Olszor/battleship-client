package main

import (
	"battleship-client/app"
	"battleship-client/http"
	"time"
)

func main() {
	client := http.NewClient("https://go-pjatk-server.fly.dev/api", time.Second*30)

	app.NewGame(client)
}
