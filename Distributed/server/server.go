package main

//go run server/server.go -broker :8050
//go run server/server.go -broker 172.31.21.113:8050
//go run ./broker/broker.go
//go test -bench=. -benchmem -run=^$

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var Before [][]uint8

func workturn1(jobs stubs.Input) []util.Cell {
	// csc stands for calculate single cell
	//this is create in old version which each logic worker can have multiple logic worker
	csc := func(i, j, width, height int) (uint8, bool) {
		sum := 0
		for m := i - 1; m <= i+1; m++ {
			for n := j - 1; n <= j+1; n++ {
				if m == i && n == j { //跳过自身
					continue
				}
				mi := m
				ni := n
				if m < 0 {
					mi = height - 1
				} else if m >= height {
					mi = 0
				}
				if n < 0 {
					ni = width - 1
				} else if n >= width {
					ni = 0
				}
				if Before[mi][ni] == 255 {
					sum++
				}
			}
		}
		var next uint8
		if Before[i][j] == 255 && (sum == 2 || sum == 3) || (Before[i][j] == 0 && sum == 3) {
			next = 255
		} else {
			next = 0
		}
		//return is return the cell value after calculate, and the boolean which decide if the cell is flip or not
		return next, next != Before[i][j]
	}

	width := len(Before[1])
	height := len(Before)
	if jobs.Thread > height {
		jobs.Thread = height
	}
	if jobs.Thread <= 0 {
		return nil
	}

	rows := jobs.End - jobs.Start

	chunk := (rows + jobs.Thread - 1) / jobs.Thread

	//wait group wait till all parallel finish
	var wg sync.WaitGroup
	if rows < jobs.Thread {
		jobs.Thread = rows
	}
	flips := make([][]util.Cell, 1024)
	start := jobs.Start
	for t := 0; t < jobs.Thread; t++ {
		//allocate the work amount each worker do
		end := start + chunk

		if end >= jobs.End {
			end = jobs.End
		}

		wg.Add(1)
		go func(idx, from, to int) {
			defer wg.Done()
			local := make([]util.Cell, 0, (to-from)*width)
			//key make it consistent
			for y := from; y < to; y++ {
				for x := 0; x < width; x++ {
					_, changed := csc(y, x, width, height)
					if changed {
						local = append(local, util.Cell{X: x, Y: y})
					}

				}
			}
			flips[idx] = local
		}(t, start, end)
		start = end
	}
	wg.Wait()

	result := make([]util.Cell, 0, 256)
	for _, slice := range flips {
		result = append(result, slice...)
	}

	return result
}

// worker is use to calculate and return the cell that next turn need to calculate
// worker and distributor using channel to communicate
func worker(jobs stubs.Input) []util.Cell {

	width := len(Before[1])
	height := len(Before)

	thisList := []util.Cell{}

	for _, val := range jobs.Thislists {
		if jobs.Start <= val.Y && val.Y < jobs.End {
			thisList = append(thisList, val)
		}
	}

	ln := len(thisList)
	if ln == 0 {
		return nil
	}

	if jobs.Thread <= 0 {
		return nil
	}
	if ln < jobs.Thread {
		jobs.Thread = ln
	}

	chunk := (ln + jobs.Thread - 1) / jobs.Thread

	flipss := make([][]util.Cell, jobs.Thread)

	var wg sync.WaitGroup
	//start for thislist
	start := 0
	for t := 0; t < jobs.Thread; t++ {

		//allocate the work amount each worker do
		end := start + chunk

		if end >= ln {
			end = ln
		}

		if start >= len(thisList) {
			break
		}
		wg.Add(1)

		go func(idx, from, to int) {
			defer wg.Done()
			flip := make([]util.Cell, 0, to-from)
			for i := from; i < to; i++ {
				cell := thisList[i]
				m := cell.Y
				n := cell.X
				sum := 0
				for a := m - 1; a <= m+1; a++ {
					for b := n - 1; b <= n+1; b++ {
						if a == m && b == n {
							continue
						}
						mi := a
						ni := b
						if a < 0 {
							mi = height - 1
						} else if a >= height {
							mi = 0
						}
						if b < 0 {
							ni = width - 1
						} else if b >= width {
							ni = 0
						}
						if Before[mi][ni] == 255 {
							sum++
						}
					}
				}
				var next uint8
				if Before[m][n] == 255 && (sum == 2 || sum == 3) || (Before[m][n] == 0 && sum == 3) {
					next = 255
				} else {
					next = 0
				}
				if Before[m][n] != next {
					flip = append(flip, util.Cell{X: n, Y: m})
				}
			}
			flipss[idx] = flip
		}(t, start, end)
		start = end
	}
	wg.Wait()

	result := make([]util.Cell, 0, 256)
	for _, val := range flipss {
		result = append(result, val...)
	}
	return result
}

type Workers struct {
}

func (w *Workers) Working(req stubs.Input, resp *stubs.WorkerResult) error {
	// log.Println("current turn  ", req.Whichturn)
	if req.Key == 'k' {
		os.Exit(0)
	}

	//broker send startrow and endrow and partial before
	var flip []util.Cell
	if req.Whichturn == 1 {
		Before = req.Before
		flip = workturn1(req)
	} else {
		if len(Before) != len(req.Before) {
			if req.Start != 0 && req.End != len(Before) {
				for i := req.Start - 1; i < req.End+1; i++ {
					copy(Before[i], req.Before[i-(req.Start-1)])
				}
			} else if req.Start == 0 {
				copy(Before[len(Before)-1], req.Before[0])
				for i := req.Start; i < req.End+1; i++ {
					copy(Before[i], req.Before[i-(req.Start-1)])
				}
			} else if req.End == len(Before) {
				copy(Before[0], req.Before[0])
				for i := req.Start - 1; i < req.End; i++ {
					copy(Before[i], req.Before[i-(req.Start-2)])
				}

			}

		} else {
			Before = req.Before
		}

		flip = worker(req)

	}
	*resp = stubs.WorkerResult{Changes: flip}
	return nil
}

func main() {
	brokerAddr := flag.String("broker", "52.201.7.150:8030", "broker rpc address")
	flag.Parse()
	w := new(Workers)
	err := rpc.Register(w)
	if err != nil {
		log.Fatalf("failed to register worker: %v", err)
	}
	//call broker rpc give server address
	conn, error := net.Dial("tcp", *brokerAddr)
	if error != nil {
		log.Fatal("fail to dial broker")
	}
	defer conn.Close()
	//serveconn make this connection to rpc receiver
	rpc.ServeConn(conn)

}
