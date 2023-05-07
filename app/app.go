package app

import (
	"battleship-client/http"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

type App struct {
	client              *http.Client
	player              string
	playerDescription   string
	opponent            string
	opponentDescription string
	ui                  *gui.GUI
	playerShips         []string
	playerBoard         *gui.Board
	opponentBoard       *gui.Board
	playerStates        [10][10]gui.State
	opponentStates      [10][10]gui.State
	opponentShots       []string
}

func NewGame(client *http.Client) App {
	err := client.InitGame()
	if err != nil {
		log.Fatal(err)
	}

	status, err := client.Status()
	if err != nil {
		log.Fatal(err)
	}
	for status.GameStatus != "game_in_progress" || !status.ShouldFire {
		time.Sleep(time.Second)
		status, err = client.Status()
		if err != nil {
			log.Fatal(err)
		}
	}

	board, err := client.Board()
	if err != nil {
		log.Fatal(err)
	}

	desc, err := client.Description()
	if err != nil {
		log.Fatal(err)
	}

	a := App{
		client:              client,
		player:              status.Nick,
		playerDescription:   desc.Desc,
		opponent:            status.Opponent,
		opponentDescription: desc.OppDesc,
		playerShips:         board,
		opponentShots:       status.OppShots,
	}

	a.render()

	return a
}

func (a *App) render() {
	a.ui = gui.NewGUI(true)

	a.playerBoard = gui.NewBoard(1, 6, nil)
	a.ui.Draw(a.playerBoard)

	a.playerStates = [10][10]gui.State{}
	for i := range a.playerStates {
		a.playerStates[i] = [10]gui.State{}
	}

	for _, coordinate := range a.playerShips {
		x, y := convertCoordinate(coordinate)
		a.playerStates[x][y] = gui.Ship
	}

	for _, coordinate := range a.opponentShots {
		x, y := convertCoordinate(coordinate)
		if a.playerStates[x][y] == gui.Ship {
			a.playerStates[x][y] = gui.Hit
		} else {
			a.playerStates[x][y] = gui.Miss
		}
	}

	a.playerBoard.SetStates(a.playerStates)

	a.opponentBoard = gui.NewBoard(46, 6, nil)
	a.ui.Draw(a.opponentBoard)

	a.opponentStates = [10][10]gui.State{}
	for i := range a.opponentStates {
		a.opponentStates[i] = [10]gui.State{}
	}

	exitTxt := gui.NewText(1, 1, "Press Ctrl+C to exit", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Black})
	a.ui.Draw(exitTxt)
	vsTxt := gui.NewText(1, 3, fmt.Sprintf("%s vs %s", a.player, a.opponent), nil)
	a.ui.Draw(vsTxt)
	for i, line := range wrapText(a.playerDescription) {
		descTxt := gui.NewText(1, 28+i, line, nil)
		a.ui.Draw(descTxt)
	}
	for i, line := range wrapText(a.opponentDescription) {
		oppDescTxt := gui.NewText(46, 28+i, line, nil)
		a.ui.Draw(oppDescTxt)
	}
	yourTurnTxt := gui.NewText(46, 3, "Your turn!", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Green})
	a.ui.Draw(yourTurnTxt)
	opponentTurnTxt := gui.NewText(46, 3, "Opponent turn!", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Red})

	go func() {
		for {
			var result string
			for result != "miss" {
				coordinate := a.opponentBoard.Listen(context.TODO())
				x, y := convertCoordinate(coordinate)
				if a.opponentStates[x][y] == gui.Hit || a.opponentStates[x][y] == gui.Miss {
					continue
				}
				fireResponse, err := a.client.Fire(coordinate)
				if err != nil {
					log.Fatal(err)
				}
				if fireResponse.Result == "miss" {
					a.opponentStates[x][y] = gui.Miss
				} else {
					a.opponentStates[x][y] = gui.Hit
				}
				a.opponentBoard.SetStates(a.opponentStates)
				result = fireResponse.Result
			}

			a.ui.Remove(yourTurnTxt)
			a.ui.Draw(opponentTurnTxt)

			status, err := a.client.Status()
			if err != nil {
				log.Fatal(err)
			}

			for !status.ShouldFire {
				a.ui.Log(strings.Join(status.OppShots, ","))
				a.playerBoard.SetStates(drawShots(a.playerStates, status.OppShots))
				if status.GameStatus == "ended" {
					var resultTxt gui.Drawable
					if status.LastGameStatus == "win" {
						resultTxt = gui.NewText(46, 3, "You won!", &gui.TextConfig{FgColor: gui.Black, BgColor: gui.Green})
					} else {
						resultTxt = gui.NewText(46, 3, "You lost!", &gui.TextConfig{FgColor: gui.Black, BgColor: gui.Red})
					}
					a.ui.Remove(opponentTurnTxt)
					a.ui.Draw(resultTxt)
					select {}
				}
				time.Sleep(time.Second)
				status, err = a.client.Status()
				if err != nil {
					log.Fatal(err)
				}
			}

			a.ui.Remove(opponentTurnTxt)
			a.ui.Draw(yourTurnTxt)
		}
	}()

	a.ui.Start(nil)
}

func convertCoordinate(coordinate string) (int, int) {
	if len(coordinate) < 2 || len(coordinate) > 3 {
		log.Fatal("coordinate should have length 2 or 3!")
	}

	if len(coordinate) == 2 {
		return int(coordinate[0]) - 65, int(coordinate[1]) - 49
	}

	return int(coordinate[0]) - 65, 9
}

func wrapText(text string) []string {
	words := strings.Split(text, " ")
	var lines []string

	line := words[0]
	for _, word := range words[1:] {
		if len(line)+len(word)+1 > 44 {
			lines = append(lines, line)
			line = word
			continue
		}
		line += " " + word
	}
	lines = append(lines, line)

	return lines
}

func drawShots(states [10][10]gui.State, shots []string) [10][10]gui.State {
	for _, coord := range shots {
		x, y := convertCoordinate(coord)
		if states[x][y] == gui.Ship {
			states[x][y] = gui.Hit
		} else {
			states[x][y] = gui.Miss
		}
	}

	return states
}
