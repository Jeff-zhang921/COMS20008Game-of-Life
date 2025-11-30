package gol

import (
	"fmt"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type Input struct {
	start       int
	end         int
	thislists   []util.Cell
	Before      [][]uint8
	x           int
	y           int
	whichturn   int
	flipsworker []util.Cell
}

type workerResult struct {
	changes []util.Cell
}

func liveCount(height, width int, Slice [][]uint8) []util.Cell {
	var liveCell []util.Cell
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if Slice[i][j] == 255 {
				liveCell = append(liveCell, util.Cell{X: j, Y: i})
			}
		}
	}
	return liveCell
}

// worker goroutine
func worker(jobs <-chan Input, result chan<- workerResult) {
	for job := range jobs {
		width := job.x
		height := job.y
		flips := job.flipsworker
		if job.whichturn == 1 {
			for i := job.start; i < job.end; i++ {
				for j := 0; j < width; j++ {
					sum := 0
					for m := i - 1; m <= i+1; m++ {
						for n := j - 1; n <= j+1; n++ {
							if m == i && n == j {
								continue
							} else {
								//not all cell need modulo
								if (0 <= m && m < height) && (0 <= n && n < width) {
									if job.Before[(m)][(n)] == 255 {
										sum++
									}
								} else {
									//don't use moduluo
									mi := m
									ni := n
									if m >= height {
										mi = 0
									} else if m < 0 {
										mi = height - 1
									}
									if n >= width {
										ni = 0
									} else if n < 0 {
										ni = width - 1
									}

									if job.Before[mi][ni] == 255 {
										sum++
									}
								}
							}
						}
					}
					var next uint8
					if job.Before[i][j] == 255 && (sum == 2 || sum == 3) || (job.Before[i][j] == 0 && sum == 3) {
						next = 255
					} else {
						next = 0
					}
					if job.Before[i][j] != next {
						flips = append(flips, util.Cell{X: j, Y: i})
					}
				}
			}
		} else {
			for _, cell := range job.thislists {
				i := cell.Y
				j := cell.X
				sum := 0
				for m := i - 1; m <= i+1; m++ {
					for n := j - 1; n <= j+1; n++ {
						if m == i && n == j {
							continue
						} else {
							if (0 <= m && m < height) && (0 <= n && n < width) {
								if job.Before[(m)][(n)] == 255 {
									sum++
								}
							} else {
								mi := m
								ni := n
								if m >= height {
									mi = 0
								} else if m < 0 {
									mi = height - 1
								}
								if n >= width {
									ni = 0
								} else if n < 0 {
									ni = width - 1
								}
								if job.Before[mi][ni] == 255 {
									sum++
								}
							}
						}
					}
				}
				var next uint8
				if job.Before[i][j] == 255 && (sum == 2 || sum == 3) || (job.Before[i][j] == 0 && sum == 3) {
					next = 255
				} else {
					next = 0
				}
				if job.Before[i][j] != next {
					flips = append(flips, util.Cell{X: j, Y: i})
				}
			}
		}
		result <- workerResult{changes: flips}
	}
}

func distributor(p Params, events chan<- Event, keyPresses <-chan rune, ioShared *IOShared) {
	quit := false
	paused := false

	x := p.ImageWidth
	y := p.ImageHeight

	ioShared.mu.Lock()
	ioShared.filename = fmt.Sprintf("%dx%d", x, y)
	ioShared.command = ioInput
	ioShared.isIdle = false
	ioShared.params.ImageHeight = y
	ioShared.params.ImageWidth = x
	ioShared.cond.Signal()

	for !ioShared.isIdle {
		ioShared.cond.Wait()
	}
	ioShared.mu.Unlock()

	Before := make([][]uint8, y)
	for i := 0; i < y; i++ {
		Before[i] = make([]uint8, x)
		copy(Before[i], ioShared.data[i*x:(i+1)*x])
	}

	initFlips := make([]util.Cell, 0)
	for i := 0; i < y; i++ {
		for j := 0; j < x; j++ {
			if Before[i][j] == 255 {
				initFlips = append(initFlips, util.Cell{X: j, Y: i})
			}
		}
	}

	if len(initFlips) > 0 {
		events <- CellsFlipped{CompletedTurns: 0, Cells: initFlips}
	}

	turn := 0
	events <- StateChange{turn, Executing}

	NextChange := make([][]bool, y)
	for i := 0; i < y; i++ {
		NextChange[i] = make([]bool, x)
	}
	//	thisList := make([]util.Cell, 0, 1024)
	nextList := make([]util.Cell, x*y)

	getNeighbor := func(ce util.Cell) {

		for a := -1; a <= 1; a++ {
			for b := -1; b <= 1; b++ {
				na, nb := a+ce.Y, b+ce.X
				if na < 0 {
					na += y
				} else if na >= y {
					na -= y
				}
				if nb < 0 {
					nb += x
				} else if nb >= x {
					nb -= x
				}
				if !NextChange[na][nb] {
					NextChange[na][nb] = true
					nextList = append(nextList, util.Cell{X: nb, Y: na})
				}
			}
		}
	}

	save := func(completed int, Slice [][]uint8) {
		out := make([]uint8, x*y)
		for i := 0; i < y; i++ {
			copy(out[i*x:(i+1)*x], Slice[i])
		}

		ioShared.mu.Lock()
		ioShared.filename = fmt.Sprintf("%dx%dx%d", x, y, completed)
		ioShared.data = out
		ioShared.command = ioOutput
		ioShared.isIdle = false
		ioShared.cond.Signal()

		for !ioShared.isIdle {
			ioShared.cond.Wait()
		}
		ioShared.mu.Unlock()

		events <- ImageOutputComplete{CompletedTurns: completed, Filename: fmt.Sprintf("%dx%dx%d", x, y, completed)}
	}

	thread := p.Threads
	jobs := make(chan Input, thread)
	results := make(chan workerResult, thread)
	flipsworker := make([][]util.Cell, thread)
	for w := 0; w < thread; w++ {
		flipsworker[w] = make([]util.Cell, 1024)
		go worker(jobs, results)
	}

	// main game logic
	flips := make([]util.Cell, 0, 1024)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for turn = 1; turn <= p.Turns; turn++ {
		flips = flips[:0]

		if turn == 1 {
			chunk := y / thread
			startRow := 0
			for w := 0; w < thread; w++ {
				endRow := startRow + chunk
				if w == thread-1 {
					endRow = y
				}
				jobs <- Input{
					start:       startRow,
					end:         endRow,
					Before:      Before,
					x:           x,
					y:           y,
					whichturn:   1,
					flipsworker: flipsworker[w],
				}
				startRow = endRow
			}
			for w := 0; w < thread; w++ {
				res := <-results
				flips = append(flips, res.changes...)
			}
			for _, cell := range flips {
				getNeighbor(cell)
			}

		} else {
			n := len(nextList)
			if n > 0 {
				chunk := n / thread
				startIndex := 0
				for w := 0; w < thread; w++ {
					flipsworker[w] = flipsworker[w][:0]
					endIndex := startIndex + chunk
					if w == thread-1 {
						endIndex = n
					}

					jobs <- Input{
						thislists:   nextList[startIndex:endIndex],
						Before:      Before,
						x:           x,
						y:           y,
						whichturn:   turn,
						flipsworker: flipsworker[w],
					}
					startIndex = endIndex
				}
				for w := 0; w < thread; w++ {
					res := <-results
					flips = append(flips, res.changes...)
				}
				nextList = nextList[:0]
				for _, cell := range flips {
					getNeighbor(cell)
				}
			}
		}

		for _, val := range nextList {
			NextChange[val.Y][val.X] = false
		}

		safe := append([]util.Cell(nil), flips...)
		events <- CellsFlipped{CompletedTurns: turn, Cells: safe}

		select {
		case COMMAND := <-keyPresses:
			switch COMMAND {
			case 's':
				save(turn-1, Before)
				events <- StateChange{turn - 1, Executing}
			case 'q':
				save(turn-1, Before)
				quit = true
				events <- FinalTurnComplete{CompletedTurns: turn - 1, Alive: liveCount(y, x, Before)}
				events <- StateChange{CompletedTurns: turn - 1, NewState: Quitting}
			case 'p':
				fmt.Println("Paused")
				paused = true
				events <- StateChange{CompletedTurns: turn - 1, NewState: Paused}
				for paused {
					key := <-keyPresses
					switch key {
					case 's':
						save(turn-1, Before)
						events <- StateChange{turn - 1, Executing}
					case 'q':
						fmt.Println("Quitting")
						save(turn-1, Before)
						quit = true
						events <- FinalTurnComplete{CompletedTurns: turn - 1, Alive: liveCount(y, x, Before)}
						events <- StateChange{CompletedTurns: turn - 1, NewState: Quitting}
						paused = false
						fmt.Println("Quitting2")
					case 'p':
						events <- StateChange{CompletedTurns: turn - 1, NewState: Executing}
						paused = false
					}
				}
			}
		default:
		}
		for _, cell := range flips {
			if Before[cell.Y][cell.X] == 255 {
				Before[cell.Y][cell.X] = 0
			} else {
				Before[cell.Y][cell.X] = 255
			}
		}
		if quit {
			break
		}

		select {
		case <-ticker.C:
			counts := 0
			for i := 0; i < y; i++ {
				for j := 0; j < x; j++ {
					if Before[i][j] == 255 {
						counts++
					}
				}
			}
			events <- AliveCellsCount{turn, counts}
		default:
		}

		events <- TurnComplete{CompletedTurns: turn}
	}

	close(jobs)

	save(p.Turns, Before)

	events <- StateChange{p.Turns, Quitting}
	events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: liveCount(y, x, Before)}
	close(events)
	
}
