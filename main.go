package main

import (
	"battleship-client/app"
	"battleship-client/http"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	trimFunc := func(c rune) bool {
		return c == '\r' || c == '\n'
	}
	fmt.Print("Enter your nick name or leave empty for random: ")
	nick, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	nick = strings.TrimRightFunc(nick, trimFunc)
	description := ""
	if nick != "" {
		fmt.Print("Enter your description: ")
		description, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		description = strings.TrimRightFunc(description, trimFunc)
	}
	fmt.Print("Do you want to play against wpbot (y/n): ")
	wpbotInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	wpbot := strings.TrimRightFunc(wpbotInput, trimFunc) == "y"
	targetNick := ""

	client := http.NewClient("https://go-pjatk-server.fly.dev/api", time.Second*30)

	if !wpbot {
		playersList, err := client.List()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("List of players waiting for game:")
		for _, player := range *playersList {
			if player.GameStatus == "waiting" {
				fmt.Println(player.Nick)
			}
		}
		fmt.Print("Enter nick of player you want to play against or leave empty if you want to start waiting: ")
		targetNick, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		targetNick = strings.TrimRightFunc(targetNick, trimFunc)
	}

	app.NewGame(client, description, nick, targetNick, wpbot)
}
