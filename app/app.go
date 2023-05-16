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
	shouldFire          bool
	yourTurnTxt         *gui.Text
	opponentTurnTxt     *gui.Text
	timerTxt            *gui.Text
	accuracyTxt         *gui.Text
	shotsFired          int
	shotsHit            int
}

func NewApp(client *http.Client) App {
	return App{
		client: client,
	}
}

func NewGame(client *http.Client, description string, nick string, targetNick string, wpbot bool) App {
	err := client.InitGame(description, nick, targetNick, wpbot)
	if err != nil {
		log.Fatal(err)
	}

	status, err := client.Status()
	if err != nil {
		log.Fatal(err)
	}

	counter := 0
	for status.GameStatus != "game_in_progress" {
		if (counter+1)%10 == 0 {
			err := client.Refresh()
			if err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		status, err = client.Status()
		if err != nil {
			log.Fatal(err)
		}
		counter++
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
		shouldFire:          status.ShouldFire,
	}

	a.run()

	return a
}

func (a *App) run() {
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
	a.yourTurnTxt = gui.NewText(46, 3, "Your turn!", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Green})
	a.opponentTurnTxt = gui.NewText(46, 3, "Opponent turn!", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Red})
	a.displayTurnInfo()
	a.timerTxt = gui.NewText(1, 2, "", nil)
	a.ui.Draw(a.timerTxt)
	a.accuracyTxt = gui.NewText(46, 1, fmt.Sprintf("Accuracy: %d/%d", a.shotsHit, a.shotsFired), nil)
	a.ui.Draw(a.accuracyTxt)

	go func() {
		for {
			a.waitForYourTurn()
			a.displayTurnInfo()
			a.handleFire()
			a.displayTurnInfo()
		}
	}()

	a.ui.Start(nil)
}

func (a *App) drawOppShots() {
	for _, coord := range a.opponentShots {
		x, y := convertCoordinate(coord)
		if a.playerStates[x][y] == gui.Hit || a.playerStates[x][y] == gui.Miss {
			continue
		}
		if a.playerStates[x][y] == gui.Ship {
			a.playerStates[x][y] = gui.Hit
		} else {
			a.playerStates[x][y] = gui.Miss
		}
	}
	a.playerBoard.SetStates(a.playerStates)
}

func (a *App) displayTurnInfo() {
	a.ui.Remove(a.yourTurnTxt)
	a.ui.Remove(a.opponentTurnTxt)
	if a.shouldFire {
		a.ui.Draw(a.yourTurnTxt)
	} else {
		a.ui.Draw(a.opponentTurnTxt)
	}
}

func (a *App) waitForYourTurn() {
	for !a.shouldFire {
		time.Sleep(time.Second)
		status, err := a.client.Status()
		if err != nil {
			log.Fatal(err)
		}

		a.shouldFire = status.ShouldFire
		a.opponentShots = status.OppShots
		a.drawOppShots()

		if status.GameStatus == "ended" {
			a.handleGameEnded(status.LastGameStatus)
		}
	}
}

func (a *App) handleFire() {
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
		a.shotsFired++
		if fireResponse.Result == "miss" {
			a.opponentStates[x][y] = gui.Miss
		} else {
			a.opponentStates[x][y] = gui.Hit
			a.shotsHit++
		}
		a.opponentBoard.SetStates(a.opponentStates)
		a.accuracyTxt.SetText(fmt.Sprintf("Accuracy: %d/%d", a.shotsHit, a.shotsFired))

		if fireResponse.Result == "sunk" {
			status, err := a.client.Status()
			if err != nil {
				log.Fatal(err)
			}

			if status.GameStatus == "ended" {
				a.handleGameEnded(status.LastGameStatus)
			}
		}

		result = fireResponse.Result
	}

	a.shouldFire = false
}

func (a *App) handleGameEnded(result string) {
	var resultTxt gui.Drawable
	if result == "win" {
		resultTxt = gui.NewText(46, 3, "You won!", &gui.TextConfig{FgColor: gui.Black, BgColor: gui.Green})
	} else {
		resultTxt = gui.NewText(46, 3, "You lost!", &gui.TextConfig{FgColor: gui.Black, BgColor: gui.Red})
	}
	a.ui.Remove(a.opponentTurnTxt)
	a.ui.Remove(a.yourTurnTxt)
	a.ui.Draw(resultTxt)
	select {}
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
