//divide by 0
//threadforworker more even

package main

import (
	"encoding/gob"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)


var handlekeyinrunout chan rune
var keyBuffer []rune
var streamListener net.Listener
var WorkerAddr []*rpc.Client

// lock is use when write to wrokeraddr
var lock sync.Mutex

// flag might modify by go function
var flaglock sync.Mutex

var runlock sync.Mutex

func allocatethreadtoworker(thread int, clients []*rpc.Client, threadperworker []int) (bool, []int, int) {
	workernode := len(clients)
	if workernode == 0 {
		return false, threadperworker, 0
	}

	if thread < workernode {
		workernode = thread
		if workernode == 0 {
			workernode = 1
		}
		for i := 0; i < workernode; i++ {
			clients[i] = WorkerAddr[i]
			threadperworker[i] = 1
		}
	} else {
		for i := 0; i < workernode; i++ {
			clients[i] = WorkerAddr[i]
		}
		total := thread
		base := total / workernode
		rem := total % workernode

		for i := 0; i < workernode; i++ {
			size := base
			if rem > 0 {
				size++
				rem--
			}
			if size < 1 {
				size = 1
			}
			threadperworker[i] = size
		}
	}
	
	return true, threadperworker, workernode
}

//calculate live cells in current world
func liveCells(world [][]uint8, x int, y int) []util.Cell {
	cells := make([]util.Cell, 0)
	for i := 0; i < y; i++ {
		for j := 0; j < x; j++ {
			if world[i][j] == 255 {
				cells = append(cells, util.Cell{X: j, Y: i})
			}
		}
	}
	return cells
}

func getNeighbor(ce util.Cell, nextChange [][]bool, nextList []util.Cell, x int, y int) []util.Cell {
	for a := -1; a <= 1; a++ {
		for b := -1; b <= 1; b++ {
			na := a + ce.Y
			nb := b + ce.X
			if (0 <= (a+ce.Y) && (a+ce.Y) < y) && (0 <= (b+ce.X) && (b+ce.X) < x) {
				if !nextChange[na][nb] {
					nextChange[na][nb] = true
					nextList = append(nextList, util.Cell{X: nb, Y: na})
				}
			} else {
				na = (na + y) % y
				nb = (nb + x) % x
				if !nextChange[na][nb] {
					nextChange[na][nb] = true
					nextList = append(nextList, util.Cell{X: nb, Y: na})
				}
			}
		}
	}
	return nextList
}

//nextlist->thislist
func rollList(thisList []util.Cell, nextList []util.Cell, nextChange [][]bool) ([]util.Cell, []util.Cell, [][]bool) {
	thisList = thisList[:0]
	thisList = append(thisList, nextList...)
	for _, val := range nextList {
		nextChange[val.Y][val.X] = false
	}
	nextList = nextList[:0]
	return thisList, nextList, nextChange
}

func turn1(y int, workernode int, threadperworker []int, Before [][]uint8, clients []*rpc.Client, results chan []util.Cell) (int, int, bool) {
	var wg sync.WaitGroup
	send := 0
	flag := false
	deadWorker := -1
	rowsPerWorker := (y + workernode - 1) / workernode

	send = 0

	for i := 0; i < workernode; i++ {
		thread := threadperworker[i]
		//allocate rows to each worker
		startRow := rowsPerWorker * i
		endRow := startRow + rowsPerWorker
		if startRow > y {
			break
		}
		if endRow > y {
			endRow = y
		}
		wg.Add(1)
		//worker only response to work on the field require
		inputting := stubs.Input{
			Start:     startRow,
			End:       endRow,
			Before:    Before,
			Whichturn: 1,
			Thread:    thread,
		}

		send++
		go func(req stubs.Input, workernodes int) {
			defer wg.Done()
			var resp stubs.WorkerResult
			if err := clients[workernodes].Call(stubs.Work, req, &resp); err != nil {
				flaglock.Lock()
				flag = true
				deadWorker = workernodes
				flaglock.Unlock()
				clients[workernodes].Close()
			} else {
				results <- resp.Changes
			}
		}(inputting, i)
	}
	wg.Wait()
	return deadWorker, send, flag
}

func turnX(y int, workernode int, threadperworker []int, Before [][]uint8, clients []*rpc.Client, results chan []util.Cell, thisList []util.Cell, turn int, topBuf, bottomBuf [][]uint8) (int, int, bool) {

	var wg sync.WaitGroup
	send := 0
	flag := false
	deadWorker := -1
	//evenly divided
	rowsPerWorker := (y + workernode - 1) / workernode

	n := len(thisList)

	if n > 0 {
		send = 0
		for i := 0; i < workernode; i++ {

			thread := threadperworker[i]

			startRow := rowsPerWorker * i
			endRow := startRow + rowsPerWorker
			if startRow >= y {
				break
			}
			if endRow >= y {
				endRow = y
			}
			var input stubs.Input
			wg.Add(1)

			if workernode == 1 {

				input = stubs.Input{
					Thislists: thisList,
					Start:     0,
					End:       len(Before),
					Before:    Before,
					Whichturn: turn,
					Thread:    thread,
				}
			} else {
				sendsRow := startRow - 1
				sendeRow := endRow + 1
				if sendsRow > 0 && sendeRow < y {
					input = stubs.Input{
						Thislists: thisList,
						Start:     startRow,
						End:       endRow,
						Before:    Before[sendsRow:sendeRow],
						Whichturn: turn,
						Thread:    thread,
					}
				}

				if sendsRow < 0 {
					rowneed := sendeRow - sendsRow
					buf := topBuf[:rowneed]
					copy(buf[0], Before[y-1])
					for i := 0; i < sendeRow; i++ {
						copy(buf[i+1], Before[i])
					}
					input = stubs.Input{
						Thislists: thisList,
						Start:     startRow,
						End:       endRow,
						Before:    buf,
						Whichturn: turn,
						Thread:    thread,
					}
				}
				if sendeRow > y {
					rowneed := sendeRow - sendsRow

					buf := bottomBuf[:rowneed]
					for i := sendsRow; i < y; i++ {
						copy(buf[i-sendsRow+1], Before[i])
					}
					copy(buf[0], Before[0])
					//	sendeRow=1
					input = stubs.Input{
						Thislists: thisList,
						Start:     startRow,
						End:       endRow,
						Before:    buf,
						Whichturn: turn,
						Thread:    thread,
					}
				}
			}

			send++
			go func(req stubs.Input, workernodes int) {
				defer wg.Done()
				var resp stubs.WorkerResult
				errors := clients[workernodes].Call(stubs.Work, req, &resp)
				
				if errors != nil {
					flaglock.Lock()
					flag = true
					deadWorker = workernodes
					flaglock.Unlock()
					// log.Println("worker ", workernodes, " new add")
					clients[workernodes].Close()
				} else {
					results <- resp.Changes
				}
			}(input, i)
		}
		wg.Wait()
	}
	return deadWorker, send, flag
}

func dealingkeypress(encode *gob.Encoder, turn int, Before [][]uint8, quit bool, clients []*rpc.Client, x int, y int, handlekeyinrunout chan rune) (error, bool) {
	var errs error

	select {
	case key := <-handlekeyinrunout:
		switch key {
		case 's':
			err := encode.Encode(stubs.Instruct{
				Name:         "COMMAND",
				Key:          's',
				CompleteTurn: turn - 1,
				Slice:        Before,
			})
			if err != nil {
				errs = err
			}
		case 'q':
			err := encode.Encode(stubs.Instruct{
				Name:         "COMMAND",
				Key:          'q',
				CompleteTurn: turn - 1,
				Slice:        Before,
				LiveCount:    liveCells(Before, x, y),
			})
			if err != nil {
				errs = err
			}
			quit = true
		case 'k':
			err := encode.Encode(stubs.Instruct{
				Name:         "COMMAND",
				Key:          'k',
				CompleteTurn: turn - 1,
				Slice:        Before,
				LiveCount:    liveCells(Before, x, y),
			})
			if err != nil {

			}
			key := stubs.Input{Key: 'k'}
			for val := range clients {
				clients[val].Go(stubs.Work, key, new(stubs.Void), nil)
				clients[val].Close()
			}

			os.Exit(0)
			quit = true
		case 'p':
			log.Println("broker doing key p")
			err := encode.Encode(stubs.Instruct{
				Name:         "COMMAND",
				Key:          'p',
				CompleteTurn: turn - 1,
			})
			if err != nil {
				errs = err
			}
			paused := true
			for paused {
				nextKey := <-handlekeyinrunout
				switch nextKey {
				case 's':
					err := encode.Encode(stubs.Instruct{
						Name:         "COMMAND",
						Key:          's',
						CompleteTurn: turn - 1,
						Slice:        Before,
					})
					if err != nil {
						errs = err
					}

				case 'q':
					err := encode.Encode(stubs.Instruct{
						Name:         "COMMAND",
						Key:          'q',
						CompleteTurn: turn - 1,
						Slice:        Before,
						LiveCount:    liveCells(Before, x, y),
					})
					if err != nil {
						errs = err
					}
					quit = true
					paused = false

				case 'k':
					err := encode.Encode(stubs.Instruct{
						Name:         "COMMAND",
						Key:          'k',
						CompleteTurn: turn - 1,
						Slice:        Before,
						LiveCount:    liveCells(Before, x, y),
					})
					if err != nil {
						log.Println(err)
					}

					key := stubs.Input{Key: 'k'}
					for val := range clients {
						clients[val].Go(stubs.Work, key, new(stubs.Void), nil)
						clients[val].Close()
					}
					os.Exit(0)
					quit = true
				case 'p':
					err := encode.Encode(stubs.Instruct{
						Name:         "COMMAND",
						Key:          'p',
						CompleteTurn: turn - 1,
					})
					if err != nil {
						errs = err
					}

					paused = false
				}
			}
		}
	default:
		break
	}
	return errs, quit
}

type Broker struct{}

// handle key rpc handle key from distributor send to run
func (b *Broker) Handlekey(key rune, _ *stubs.Void) error {
	//if run didn't start, keybuffer keep the key and after run start give run
	if handlekeyinrunout == nil {
		keyBuffer = append(keyBuffer, key)
		return nil
	}

	//put key to the channel
	handlekeyinrunout <- key
	return nil
}

// main logic
// Run first listen to TCP event, send tcp event
// for each turn, it divide work,send corresponding workload and thread to each workernode
// with dynamic dispatch the most even work to each workernode
// if a workernode dead, it will set flag to true no matter which dead, and reallocate work to alive worker, rerun this turn
// in server, for one workernode, it receive the start of job and end of job, and the thread it will use to
// allocate each logic worker, with most even work.
func (b *Broker) Run(request stubs.BrokerInput, response *stubs.BrokerInit) error {

	conn, err := streamListener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	encode := gob.NewEncoder(conn)

	x := request.Width
	y := request.Height

	Before := make([][]uint8, y)

	for i := 0; i < y; i++ {
		Before[i] = make([]uint8, x)
		copy(Before[i], request.Data[i*x:(i+1)*x])
	}

	runlock.Lock()
	defer runlock.Unlock()

	keys := make(chan rune, 64)
	handlekeyinrunout = keys
	for _, k := range keyBuffer {
		handlekeyinrunout <- k
	}
	keyBuffer = keyBuffer[:0]

	nextChange := make([][]bool, y)
	for i := 0; i < y; i++ {
		nextChange[i] = make([]bool, x)
	}
	thisList := make([]util.Cell, 0, 1024)
	nextList := make([]util.Cell, 0, 1024)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	//call once instead of mutiple time
	//allocate how many to call and size of thread for each

	currentturn := make(chan int, 2)
	//the dead worker at nth position of workeraddr
	var deadWorker = -1

	quit := false

	log.Println("broker started run")

Workerfail:
// log.Println("worker fail handling at turn")
	results := make(chan []util.Cell, 1024)
	if deadWorker != -1 {
		WorkerAddr = append(WorkerAddr[:deadWorker], WorkerAddr[deadWorker+1:]...)
	}
	lens := len(WorkerAddr)

	clients := make([]*rpc.Client, lens)

	var workernode int
	threadperworker := make([]int, lens)
	success, threadperworker, workernode := allocatethreadtoworker(request.Thread, clients, threadperworker)

	if !success {
		log.Println("unsuccessful allocate thread to worker")
		goto Workerfail
	}
	
	rowsPerWorker := (y + workernode - 1) / workernode
	rowneed := rowsPerWorker + 2
	topBuf := make([][]uint8, rowneed)
	bottomBuf := make([][]uint8, rowneed)
	for i := range topBuf {
		topBuf[i] = make([]uint8, x)
		bottomBuf[i] = make([]uint8, x)
	}

	for turn := 1; turn <= request.Turns; turn++ {
		select {
		case current := <-currentturn:
			// Channel has output
			turn = current
		default:
		}

		flips := make([]util.Cell, 0, 256)
		flag := false
		var send int
		if turn == 1 {
			deadWorker, send, flag = turn1(y, workernode, threadperworker, Before, clients, results)
		} else {
			deadWorker, send, flag = turnX(y, workernode, threadperworker, Before, clients, results, thisList, turn, topBuf, bottomBuf)
		}
		flips = flips[:0]

		//every turn do
		if flag {
			currentturn <- turn
			for len(results) > 0 {
				<-results
			}
			log.Println("worker fail,redo turn ", turn)
			goto Workerfail
		}

		for i := 0; i < send; i++ {
			res := <-results
			flips = append(flips, res...)
		}

		for _, cell := range flips {
			nextList = getNeighbor(cell, nextChange, nextList, x, y)
		}

		thisList, nextList, nextChange = rollList(thisList, nextList, nextChange)
		err := encode.Encode(stubs.Instruct{
			Name:         "TURNEND",
			CompleteTurn: turn,
			Cells:        flips,
		})
		if err != nil {
			return err
		}

		var errors error
		errors, quit = dealingkeypress(encode, turn, Before, quit, clients, x, y, handlekeyinrunout)

		if errors != nil {
			return err
		}

		if quit {
			break
		}

		//update world
		for _, cell := range flips {
			if Before[cell.Y][cell.X] == 255 {
				Before[cell.Y][cell.X] = 0
			} else {
				Before[cell.Y][cell.X] = 255
			}
		}

		select {
		case <-ticker.C:

			count := 0
			for i := 0; i < y; i++ {
				for j := 0; j < x; j++ {
					if Before[i][j] == 255 {
						count++
					}
				}
			}
			err := encode.Encode(stubs.Instruct{
				Name:         "ALIVE",
				CompleteTurn: turn,
				CellCount:    count,
			})
			if err != nil {
				return nil
			}
		default:
		}

	}

	out := make([]uint8, x*y)
	for i := 0; i < y; i++ {
		copy(out[i*x:(i+1)*x], Before[i])
	}

	errors := encode.Encode(stubs.Instruct{
		Name:         "FINAL",
		CompleteTurn: request.Turns,
		LiveCount:    liveCells(Before, x, y),
		Out:          out,
	})
	if errors != nil {
		handlekeyinrunout = nil
		keyBuffer = keyBuffer[:0]
		return err
	}

	//when every turn finish make sure all is close cleanly
	handlekeyinrunout = nil
	keyBuffer = keyBuffer[:0]
	return nil
}

func main() {
	ln, err := net.Listen("tcp", ":8032")
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", ":8032", err)
	}
	streamListener = ln

	//register rpc
	Brokers := new(Broker)
	errros := rpc.Register(Brokers)
	if errros != nil {
		log.Fatalf("failed to register broker RPC: %v", err)
	}
	//listen to distributor rpc call
	listener1, err := net.Listen("tcp", ":8030")
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", ":8030", err)
	}
	defer listener1.Close()
	go rpc.Accept(listener1)

	//listen to worker connect
	listener2, err := net.Listen("tcp", ":8050")
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", ":8050", err)
	}
	defer listener2.Close()

	for {
		lock.Lock()
		conn, err := listener2.Accept()
		log.Println("[Broker]: Worker connected:", conn.RemoteAddr().String())
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		//upgrade to rpc client
	client := rpc.NewClient(conn)
	WorkerAddr = append(WorkerAddr, client)
	   lock.Unlock()
	}	
}
