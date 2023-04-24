package app

import (
	gui "github.com/grupawp/warships-gui/v2"
	"log"
)

func Render(board []string) {

	ui := gui.NewGUI(true)

	b := gui.NewBoard(1, 2, nil)
	ui.Draw(b)

	states := [10][10]gui.State{}
	for i := range states {
		states[i] = [10]gui.State{}
	}

	for _, coordinate := range board {
		x, y := convertCoordinate(coordinate)
		states[x][y] = gui.Ship
	}

	b.SetStates(states)

	ui.Start(nil)
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
