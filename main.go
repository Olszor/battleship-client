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

	client := http.NewClient("https://go-pjatk-server.fly.dev/api", time.Second*30)

	app.NewApp(client, reader, trimFunc, description, nick)
}
