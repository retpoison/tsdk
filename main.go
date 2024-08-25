package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/retpoison/sug"
)

type Game struct {
	s           tcell.Screen
	styles      map[string]tcell.Style
	sdk         *sug.Sudoku
	cPuzzle     [][]int
	sdkrow      int
	sdkcol      int
	screenrow   int
	screencol   int
	mainTable   [][]rune
	stopTimer   chan bool
	timeStrChan chan string
}

func main() {
	var game = Game{}
	var err error
	game.s, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err = game.s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	game.styles = map[string]tcell.Style{
		"empty": tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlack),
		"filled": tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorSlateGray),
		"selected": tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlue),
		"selectedBW": tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorWhite),
	}
	game.stopTimer = make(chan bool)
	game.timeStrChan = make(chan string)

	defer func() {
		r := recover()
		game.s.Fini()
		if r != nil {
			panic(r)
		}
	}()

	if !game.setSudoku() {
		return
	}
	game.cPuzzle = sug.CopySudoku(game.sdk.Puzzle)
	game.mainTable = game.makeTable()
	game.drawTable()
	game.run()
}
