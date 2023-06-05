package app

import (
	"battleship-client/http"
	"bufio"
	"context"
	"fmt"
	"log"
	"strconv"
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
	timer               int
	accuracyTxt         *gui.Text
	shotsFired          int
	shotsHit            int
	lastGameStatus      string
	cancelFunc          func()
	opponentShips       map[int]int
	shipsInfoTxt        *gui.Text
	customShips         []string
}

func NewApp(client *http.Client, reader *bufio.Reader, trimFunc func(rune) bool, description string, nick string) {
	a := App{
		client:            client,
		player:            nick,
		playerDescription: description,
	}

	for {
		targetNick, wpbot := a.displayMenu(reader, trimFunc)
		a.newGame(a.playerDescription, a.player, targetNick, wpbot)
		if a.lastGameStatus == "" {
			var err error
			makeRequest(func() error {
				err = a.client.Abandon()
				return err
			})
		}
	}
}

func (a *App) displayMenu(reader *bufio.Reader, trimFunc func(rune) bool) (string, bool) {
	options := []string{
		"Play against wpbot",
		"Play against one of currently waiting players",
		"Join waiting list",
		"Display top 10 players stats",
		"Display your stats",
		"Setup your board",
	}

	for {
		switch getChoice(reader, trimFunc, options) {
		case 1:
			return "", true
		case 2:
			var playersList *[]http.ListResponse
			var err error
			makeRequest(func() error {
				playersList, err = a.client.List()
				return err
			})
			if err != nil {
				log.Fatal(err)
			}
			filteredList := Filter(*playersList, func(element http.ListResponse) bool {
				return element.GameStatus == "waiting"
			})
			mappedList := Map(filteredList, func(element http.ListResponse) string {
				return element.Nick
			})
			chosenPlayerId := getChoice(reader, trimFunc, mappedList)
			return mappedList[chosenPlayerId-1], false
		case 3:
			return "", false
		case 4:
			a.displayStats()
		case 5:
			a.displayPlayerStats()
		case 6:
			a.setupBoard()
			fmt.Println(a.customShips)
		}
	}
}

func (a *App) newGame(description string, nick string, targetNick string, wpbot bool) {
	a.reset()
	var err error
	makeRequest(func() error {
		err = a.client.InitGame(a.customShips, description, nick, targetNick, wpbot)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	var status *http.StatusResponse
	makeRequest(func() error {
		status, err = a.client.Status()
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	counter := 0
	for status.GameStatus != "game_in_progress" {
		if (counter+1)%10 == 0 {
			fmt.Println("waiting...")
			makeRequest(func() error {
				err := a.client.Refresh()
				return err
			})
			if err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		makeRequest(func() error {
			status, err = a.client.Status()
			return err
		})
		if err != nil {
			log.Fatal(err)
		}
		counter++
	}
	var board []string
	makeRequest(func() error {
		board, err = a.client.Board()
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	var desc *http.DescriptionResponse
	makeRequest(func() error {
		desc, err = a.client.Description()
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	a.player = status.Nick
	a.playerDescription = desc.Desc
	a.opponent = status.Opponent
	a.opponentDescription = desc.OppDesc
	a.playerShips = board
	a.opponentShots = status.OppShots
	a.shouldFire = status.ShouldFire
	a.timer = status.Timer

	a.run()
}

func (a *App) reset() {
	a.shotsHit = 0
	a.shotsFired = 0
	a.lastGameStatus = ""
	a.timer = 0
	a.opponentShips = map[int]int{
		4: 1,
		3: 2,
		2: 3,
		1: 4,
	}
}

func (a *App) run() {
	a.ui = gui.NewGUI(true)

	a.playerBoard = gui.NewBoard(1, 8, nil)
	a.ui.Draw(a.playerBoard)

	a.playerStates = [10][10]gui.State{}
	a.opponentStates = [10][10]gui.State{}
	for i := range a.playerStates {
		a.playerStates[i] = [10]gui.State{}
		a.opponentStates[i] = [10]gui.State{}

		for j := range a.playerStates[i] {
			a.playerStates[i][j] = gui.Empty
			a.opponentStates[i][j] = gui.Empty
		}
	}

	for _, coordinate := range a.playerShips {
		x, y := convertCoordinate(coordinate)
		a.playerStates[x][y] = gui.Ship
	}

	a.playerBoard.SetStates(a.playerStates)

	a.opponentBoard = gui.NewBoard(46, 8, nil)
	a.ui.Draw(a.opponentBoard)

	exitTxt := gui.NewText(1, 1, "Press Ctrl+C to exit", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Black})
	a.ui.Draw(exitTxt)
	vsTxt := gui.NewText(1, 5, fmt.Sprintf("%s vs %s", a.player, a.opponent), nil)
	a.ui.Draw(vsTxt)
	for i, line := range wrapText(a.playerDescription) {
		descTxt := gui.NewText(1, 30+i, line, nil)
		a.ui.Draw(descTxt)
	}
	for i, line := range wrapText(a.opponentDescription) {
		oppDescTxt := gui.NewText(46, 30+i, line, nil)
		a.ui.Draw(oppDescTxt)
	}
	a.yourTurnTxt = gui.NewText(46, 5, "Your turn!", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Green})
	a.opponentTurnTxt = gui.NewText(46, 5, "Opponent turn!", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Red})
	a.displayTurnInfo()
	a.timerTxt = gui.NewText(46, 3, "", nil)
	a.ui.Draw(a.timerTxt)
	a.accuracyTxt = gui.NewText(46, 1, fmt.Sprintf("Accuracy: %d/%d", a.shotsHit, a.shotsFired), nil)
	a.ui.Draw(a.accuracyTxt)
	a.shipsInfoTxt = gui.NewText(1, 3, fmt.Sprintf("Remaining ships: 4-%d 3-%d 2-%d 1-%d", a.opponentShips[4], a.opponentShips[3], a.opponentShips[2], a.opponentShips[1]), nil)
	a.ui.Draw(a.shipsInfoTxt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context) {
		a.waitForYourTurn()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				a.displayTurnInfo()
				a.handleFire(ctx)
				a.displayTurnInfo()
				a.waitForYourTurn()
			}
		}
	}(ctx)

	a.cancelFunc = cancel

	a.ui.Start(context.Background(), nil)
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
		var status *http.StatusResponse
		var err error
		makeRequest(func() error {
			status, err = a.client.Status()
			return err
		})
		if err != nil {
			return
		}

		a.shouldFire = status.ShouldFire
		a.opponentShots = status.OppShots
		a.drawOppShots()
		a.timer = status.Timer

		if status.GameStatus == "ended" {
			a.lastGameStatus = status.LastGameStatus
			a.handleGameEnded()
		}
	}
}

func (a *App) handleFire(mainCtx context.Context) {
	var result string
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.resetTimer(&a.timer, ctx)
	for result != "miss" {
		coordinate := a.opponentBoard.Listen(mainCtx)
		if coordinate == "" {
			break
		}
		x, y := convertCoordinate(coordinate)
		if a.opponentStates[x][y] == gui.Hit || a.opponentStates[x][y] == gui.Miss {
			continue
		}
		var fireResponse *http.FireResponse
		var err error
		makeRequest(func() error {
			fireResponse, err = a.client.Fire(coordinate)
			return err
		})
		if err != nil {
			break
		}
		a.shotsFired++
		if fireResponse.Result == "miss" {
			a.opponentStates[x][y] = gui.Miss
		} else {
			a.opponentStates[x][y] = gui.Hit
			a.shotsHit++
			a.timerTxt.SetText("60")
			a.timer = 60
			if fireResponse.Result == "sunk" {
				a.handleSunk(coordinate)
			}
		}
		a.opponentBoard.SetStates(a.opponentStates)
		a.accuracyTxt.SetText(fmt.Sprintf("Accuracy: %d/%d", a.shotsHit, a.shotsFired))

		if fireResponse.Result == "sunk" {

			var status *http.StatusResponse
			makeRequest(func() error {
				status, err = a.client.Status()
				return err
			})
			if err != nil {
				log.Fatal(err)
			}

			if status.GameStatus == "ended" {
				a.lastGameStatus = status.LastGameStatus
				cancel()
				a.handleGameEnded()
			}
		}

		result = fireResponse.Result
	}

	cancel()
	a.shouldFire = false
}

func (a *App) handleGameEnded() {
	var resultTxt gui.Drawable
	if a.lastGameStatus == "win" {
		resultTxt = gui.NewText(46, 3, "You won!", &gui.TextConfig{FgColor: gui.Black, BgColor: gui.Green})
	} else {
		resultTxt = gui.NewText(46, 3, "You lost!", &gui.TextConfig{FgColor: gui.Black, BgColor: gui.Red})
	}
	a.ui.Remove(a.opponentTurnTxt)
	a.ui.Remove(a.yourTurnTxt)
	a.ui.Draw(resultTxt)
	a.cancelFunc()
}

func (a *App) displayStats() {
	var stats *http.StatsResponse
	var err error
	makeRequest(func() error {
		stats, err = a.client.Stats()
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Printf("| %s | %-20s | %s | %s | %s |\n", "RANK", "NICK", "GAMES", "WINS", "POINTS")
	for _, s := range stats.Stats {
		fmt.Printf("| %4d | %-20s | %5d | %4d | %6d |\n",
			s.Rank,
			s.Nick,
			s.Games,
			s.Wins,
			s.Points,
		)
	}
	fmt.Println()
}

func (a *App) displayPlayerStats() {
	var stats *http.PlayerStatsResponse
	var err error
	makeRequest(func() error {
		stats, err = a.client.PlayerStats(a.player)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Printf("| %s | %-20s | %s | %s | %s |\n", "RANK", "NICK", "GAMES", "WINS", "POINTS")
	fmt.Printf("| %4d | %-20s | %5d | %4d | %6d |\n",
		stats.Stats.Rank,
		stats.Stats.Nick,
		stats.Stats.Games,
		stats.Stats.Wins,
		stats.Stats.Points,
	)

	fmt.Println()
}

func (a *App) resetTimer(timer *int, ctx context.Context) {
	if timer != nil {
		a.timer = *timer
	} else {
		a.timer = 60
	}
	a.timerTxt.SetText(strconv.Itoa(a.timer))
	go func(context context.Context) {
		for {
			select {
			case <-context.Done():
				a.timerTxt.SetText("")
				return
			default:
				time.Sleep(time.Second)
				if a.timer > 0 {
					a.timer--
				}
				a.timerTxt.SetText(strconv.Itoa(a.timer))
			}
		}
	}(ctx)

}

func (a *App) handleSunk(coord string) {
	ship := getShip(a.opponentStates, coord)
	setImpossiblePositions(&a.opponentStates, ship)
	a.opponentShips[len(ship)]--
	a.shipsInfoTxt.SetText(fmt.Sprintf("Remaining ships: 4-%d 3-%d 2-%d 1-%d", a.opponentShips[4], a.opponentShips[3], a.opponentShips[2], a.opponentShips[1]))
}

func (a *App) setupBoard() {
	ships := map[int]int{
		4: 1,
		3: 2,
		2: 3,
		1: 4,
	}
	var shipsCoord []string
	ui := gui.NewGUI(true)
	board := gui.NewBoard(1, 8, nil)
	states := [10][10]gui.State{}
	for i := range states {
		states[i] = [10]gui.State{}
		for j := range states[i] {
			states[i][j] = gui.Empty
		}
	}
	ui.Draw(board)
	board.SetStates(states)
	placedShipTxt := gui.NewText(1, 3, "", nil)
	ui.Draw(placedShipTxt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for shipLength := 4; shipLength > 0; shipLength-- {
			shipsCount := ships[shipLength]
			for i := 0; i < shipsCount; i++ {
				var ship []string
				placedShipTxt.SetText(fmt.Sprintf("Place ship of length %d", shipLength))

				for i := range states {
					for j := range states[i] {
						if states[i][j] == gui.Empty {
							states[i][j] = gui.Hit
						}
					}
				}
				board.SetStates(states)

				for i := 0; i < shipLength; i++ {
					for {
						coord := board.Listen(ctx)
						if coord == "" {
							return
						}
						x, y := convertCoordinate(coord)

						if states[x][y] == gui.Hit {
							ship = append(ship, coord)
							states[x][y] = gui.Ship
							for j := range states {
								for k := range states[i] {
									if states[j][k] == gui.Hit {
										states[j][k] = gui.Empty
									}
								}
							}
							setPossiblePositions(&states, ship)
							board.SetStates(states)
							break
						}
					}
				}

				for i := range states {
					for j := range states[i] {
						if states[i][j] == gui.Hit {
							states[i][j] = gui.Empty
						}
					}
				}
				var shipPoints []point
				for _, coord := range ship {
					x, y := convertCoordinate(coord)
					shipPoints = append(shipPoints, point{x, y})
				}
				shipsCoord = append(shipsCoord, ship...)
				setImpossiblePositions(&states, shipPoints)
				board.SetStates(states)
			}
		}
		a.customShips = shipsCoord
		placedShipTxt.SetText("Press Ctrl + C to exit")
	}()

	ui.Start(ctx, nil)
}

type point struct {
	x int
	y int
}

func getShip(board [10][10]gui.State, coord string) []point {
	var ship []point
	x, y := convertCoordinate(coord)
	toVisit := []point{{x, y}}
	visited := make(map[point]struct{})

	offsets := []point{
		{1, 0},
		{0, 1},
		{-1, 0},
		{0, -1},
	}

	for len(toVisit) > 0 {
		current := toVisit[0]
		toVisit = toVisit[1:]
		if board[current.x][current.y] == gui.Hit {
			for _, offset := range offsets {
				checked := point{current.x + offset.x, current.y + offset.y}
				if _, exists := visited[checked]; exists || checked.x < 0 || checked.x >= 10 || checked.y < 0 || checked.y >= 10 {
					continue
				}
				toVisit = append(toVisit, checked)
			}
			ship = append(ship, current)
		}
		visited[current] = struct{}{}
	}
	return ship
}

func setImpossiblePositions(board *[10][10]gui.State, ship []point) {
	offsets := []point{
		{1, 0},
		{0, 1},
		{1, 1},
		{-1, 0},
		{0, -1},
		{-1, -1},
		{1, -1},
		{-1, 1},
	}

	for _, current := range ship {
		for _, offset := range offsets {
			checked := point{current.x + offset.x, current.y + offset.y}
			if checked.x < 0 || checked.x >= 10 || checked.y < 0 || checked.y >= 10 {
				continue
			}
			if board[checked.x][checked.y] == gui.Empty {
				board[checked.x][checked.y] = gui.Miss
			}
		}
	}
}

func setPossiblePositions(board *[10][10]gui.State, ship []string) {
	offsets := []point{
		{1, 0},
		{0, 1},
		{-1, 0},
		{0, -1},
	}

	for _, coord := range ship {
		x, y := convertCoordinate(coord)
		for _, offset := range offsets {
			checked := point{x + offset.x, y + offset.y}
			if checked.x < 0 || checked.x >= 10 || checked.y < 0 || checked.y >= 10 {
				continue
			}
			if board[checked.x][checked.y] == gui.Empty {
				board[checked.x][checked.y] = gui.Hit
			}
		}
	}
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

func getChoice(reader *bufio.Reader, trimFunc func(rune) bool, options []string) int {
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}

	for {
		choiceInput, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		choice, err := strconv.Atoi(strings.TrimRightFunc(choiceInput, trimFunc))
		if err != nil {
			fmt.Println("Your choice must be an integer number! Try again")
			continue
		}
		if choice < 1 || choice > len(options) {
			fmt.Println("You must choose from existing options! Try again")
			continue
		}
		return choice
	}
}
