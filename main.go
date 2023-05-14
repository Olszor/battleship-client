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
	fmt.Print("Enter your nick name: ")
	nick, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	nick = strings.TrimRightFunc(nick, trimFunc)
	fmt.Print("Enter your description: ")
	description, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	description = strings.TrimRightFunc(description, trimFunc)
	fmt.Print("Do you want to play against wpbot (y/n): ")
	wpbotInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	wpbot := strings.TrimRightFunc(wpbotInput, trimFunc) == "y"
	targetNick := ""
	if !wpbot {
		fmt.Print("Enter nick of player you want to play against: ")
		targetNick, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		targetNick = strings.TrimRightFunc(targetNick, trimFunc)
	}

	client := http.NewClient("https://go-pjatk-server.fly.dev/api", time.Second*30)

	app.NewGame(client, description, nick, targetNick, wpbot)
}
