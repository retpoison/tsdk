package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/retpoison/sug"
)

func (g *Game) run() {
	go g.timer()
	var content rune
	for {
		content, _, _, _ = g.s.GetContent(g.screencol, g.screenrow)
		g.s.SetContent(g.screencol, g.screenrow, content,
			nil, g.styles["selected"])

		g.s.Show()
		ev := g.s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.s.Sync()
			g.drawTable()
		case *tcell.EventKey:
			key := ev.Key()
			char := ev.Rune()
			if key == tcell.KeyEscape ||
				key == tcell.KeyCtrlC ||
				char == 'q' || char == 'Q' {
				return
			} else if char > '0' && char <= '9' {
				g.insertNumber(char)
			} else if char == ' ' ||
				key == tcell.KeyBackspace ||
				key == tcell.KeyBackspace2 {
				g.cleanCell()
			} else if key >= tcell.KeyUp && key <= tcell.KeyLeft {
				g.move(key, content)
			} else if char == 'u' || char == 'U' {
				g.undo()
			}
		}

		if g.isPuzzleSolved() {
			g.solvedScreen()
			return
		}
	}
}

func (g *Game) move(key tcell.Key, content rune) {
	if g.isSdkCellEmpty() {
		g.s.SetContent(g.screencol, g.screenrow, content,
			nil, g.styles["empty"])
	} else {
		g.s.SetContent(g.screencol, g.screenrow, content,
			nil, g.styles["filled"])
	}
	if key == tcell.KeyUp && g.sdkrow > 0 {
		g.sdkrow--
	} else if key == tcell.KeyDown && g.sdkrow < sug.Row-1 {
		g.sdkrow++
	} else if key == tcell.KeyLeft && g.sdkcol > 0 {
		g.sdkcol--
	} else if key == tcell.KeyRight && g.sdkcol < sug.Col-1 {
		g.sdkcol++
	}
	g.screenrow, g.screencol = g.makeRowCol(g.sdkrow, g.sdkcol)
}

func (g *Game) setSudoku() bool {
	var difficulty int
	switch g.getDifficulty() {
	case 0:
		difficulty = sug.Easy
	case 1:
		difficulty = sug.Medium
	case 2:
		difficulty = sug.Hard
	case 3:
		difficulty = sug.Expert
	case -1:
		return false
	}
	g.s.Clear()
	g.sdk = sug.NewSudoku(difficulty)
	return true
}

func (g *Game) drawTable() {
	var table = g.makeTable()
	for r := range len(table) {
		for c := range len(table[r]) {
			m := table[r][c] - '0'
			if (m > 10 || m < 1) ||
				(table[r][c] != '_' && g.mainTable[r][c] == '_') {
				g.s.SetContent(c, r, table[r][c],
					nil, g.styles["empty"])
			} else {
				g.s.SetContent(c, r, g.mainTable[r][c],
					nil, g.styles["filled"])
			}
		}
	}
}

func (g *Game) makeTable() [][]rune {
	var tableh, tablew = 11, 21
	var table = make([][]rune, tableh)
	for i := range tableh {
		table[i] = make([]rune, tablew)
	}
	var si, sj = 0, 0
	for i := range tableh {
		sj = 0
		for j := range tablew {
			if i == 3 || i == 7 {
				table[i][j] = tcell.RuneHLine
				continue
			} else if j%2 == 1 {
				table[i][j] = ' '
				continue
			} else if j == 6 || j == 14 {
				table[i][j] = tcell.RuneVLine
				continue
			}
			if g.cPuzzle[si][sj] == 0 {
				table[i][j] = '_'
			} else {
				table[i][j] = rune(g.cPuzzle[si][sj]) + '0'
			}
			sj++
		}
		if i != 3 && i != 7 {
			si++
		}
	}
	return table
}

func (g *Game) makeRowCol(sdkrow, sdkcol int) (int, int) {
	screenrow := int(sdkrow/3) + sdkrow
	screencol := (int(sdkcol/3) + sdkcol) * 2
	return screenrow, screencol
}

func (g *Game) getDifficulty() int {
	difficulties := []string{"Easy", "Medium", "Hard", "Expert"}
	selected := 0
	var drawOptions = func() {
		for i, v := range difficulties {
			for j, ch := range v {
				if selected == i {
					g.s.SetContent(j, i*2, ch,
						nil, g.styles["selectedBW"])
				} else {
					g.s.SetContent(j, i*2, ch,
						nil, g.styles["empty"])
				}
			}
		}
	}
	drawOptions()
	for {
		g.s.Show()
		ev := g.s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.s.Sync()
			drawOptions()
		case *tcell.EventKey:
			key := ev.Key()
			char := ev.Rune()
			if key == tcell.KeyEscape ||
				key == tcell.KeyCtrlC ||
				char == 'q' {
				return -1
			} else if key == tcell.KeyEnter {
				return selected
			} else if key == tcell.KeyDown && selected < len(difficulties)-1 {
				selected++
				drawOptions()
			} else if key == tcell.KeyUp && selected > 0 {
				selected--
				drawOptions()
			}
		}
	}
	return selected
}

func (g *Game) isSdkCellEmpty() bool {
	if g.sdk.Puzzle[g.sdkrow][g.sdkcol] == 0 {
		return true
	}
	return false
}

func (g *Game) cleanCell() {
	if !g.isSdkCellEmpty() {
		return
	}
	g.s.SetContent(g.screencol, g.screenrow, '_',
		nil, g.styles["empty"])
	g.cPuzzle[g.sdkrow][g.sdkcol] = 0
}

func (g *Game) insertNumber(num rune) {
	if !g.isSdkCellEmpty() {
		return
	}
	content, _, _, _ := g.s.GetContent(g.screencol, g.screenrow)
	g.s.SetContent(g.screencol, g.screenrow, num,
		nil, g.styles["empty"])
	g.cPuzzle[g.sdkrow][g.sdkcol] = int(num - '0')
	g.stack = g.stack.Push(Element{
		Row:  g.sdkrow,
		Col:  g.sdkcol,
		Char: content,
	})
}

func (g *Game) undo() {
	var e Element
	g.stack, e = g.stack.Pop()
	if e.Row == -1 && e.Col == -1 {
		return
	}
	scrow, sccol := g.makeRowCol(e.Row, e.Col)
	g.s.SetContent(sccol, scrow, e.Char,
		nil, g.styles["empty"])
	if e.Char == '_' {
		g.cPuzzle[e.Row][e.Col] = '0'
	} else {
		g.cPuzzle[e.Row][e.Col] = int(e.Char - '0')
	}
}

func (g *Game) timer() {
	start := time.Now()
	ticker := time.NewTicker(time.Second)
	var timeStr string
	for {
		select {
		case <-g.stopTimer:
			ticker.Stop()
			g.timeStrChan <- timeStr
			return
		case <-ticker.C:
			sinceSec := int(time.Since(start).Seconds())
			timeStr = fmt.Sprintf("%02d:%02d",
				sinceSec/60, sinceSec%60)
			for i, v := range []rune(timeStr) {
				g.s.SetContent(i, 12, v,
					nil, g.styles["empty"])
			}
			g.s.Show()
		}
	}
}

func (g *Game) isPuzzleSolved() bool {
	for i := range sug.Row {
		for j := range sug.Col {
			if g.sdk.Answers[0][i][j] != g.cPuzzle[i][j] ||
				g.cPuzzle[i][j] == 0 {
				return false
			}
		}
	}
	return true
}

func (g *Game) solvedScreen() {
	g.stopTimer <- true
	timeStr := <-g.timeStrChan
	g.s.Clear()

	solvedStr := []string{
		fmt.Sprintf("You solved the puzzle in %s.", timeStr),
		"",
		"Press any key to exit."}
	for i, s := range solvedStr {
		for j, v := range []rune(s) {
			g.s.SetContent(j, i, v, nil, g.styles["selected"])
		}
	}
	g.s.Show()
	for {
		_ = g.s.PollEvent()
		return
	}
}
