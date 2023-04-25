package app

import (
	"battleship-client/http"
	"context"
	"fmt"
	"log"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

type App struct {
	player         string
	opponent       string
	ui             *gui.GUI
	playerShips    []string
	playerBoard    *gui.Board
	opponentBoard  *gui.Board
	playerStates   [10][10]gui.State
	opponentStates [10][10]gui.State
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
	for status.GameStatus != "game_in_progress" {
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

	a := App{
		player:      status.Nick,
		opponent:    status.Opponent,
		playerShips: board,
	}

	a.render()

	return a
}

func (a *App) render() {
	a.ui = gui.NewGUI(true)

	a.playerBoard = gui.NewBoard(1, 10, nil)
	a.ui.Draw(a.playerBoard)

	a.playerStates = [10][10]gui.State{}
	for i := range a.playerStates {
		a.playerStates[i] = [10]gui.State{}
	}

	for _, coordinate := range a.playerShips {
		x, y := convertCoordinate(coordinate)
		a.playerStates[x][y] = gui.Ship
	}

	a.playerBoard.SetStates(a.playerStates)

	a.opponentBoard = gui.NewBoard(46, 10, nil)
	a.ui.Draw(a.opponentBoard)

	a.opponentStates = [10][10]gui.State{}
	for i := range a.opponentStates {
		a.opponentStates[i] = [10]gui.State{}
	}

	exitTxt := gui.NewText(1, 1, "Press Ctrl+C to exit", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Black})
	a.ui.Draw(exitTxt)
	vsTxt := gui.NewText(1, 3, fmt.Sprintf("%s vs %s", a.player, a.opponent), nil)
	a.ui.Draw(vsTxt)

	go func() {
		for {
			coordinate := a.opponentBoard.Listen(context.TODO())
			x, y := convertCoordinate(coordinate)
			a.opponentStates[x][y] = gui.Miss
			a.opponentBoard.SetStates(a.opponentStates)
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
