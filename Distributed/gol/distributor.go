package gol

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- []uint8
	ioInput    <-chan []uint8
	keyPresses <-chan rune
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	BrokerAddr := "52.201.7.150:8030"

	x := p.ImageWidth
	y := p.ImageHeight

	//let io know what to do and give the filename
	c.ioCommand <- ioInput
	c.ioFilename <- fmt.Sprintf("%dx%d", x, y)

	//imageheight which means there are how many lines
	//first create
	Before := make([][]uint8, y)

	for i := 0; i < y; i++ {
		Before[i] = make([]uint8, x)
	}

	//get the array modify io
	get := <-c.ioInput
	for i := 0; i < y; i++ {
		copy(Before[i], get[i*x:(i+1)*x])
	}

	//listen event start
	//event have listen put in to this channel
	msgs := make(chan stubs.Instruct)
	//listen from broker sending event

	//dial init first
	client, err := rpc.Dial("tcp", BrokerAddr)

	if err != nil {
		log.Printf("connect fail")
		return
	}
	defer client.Close()
	args := stubs.BrokerInput{
		Width:  x,
		Height: y,
		Turns:  p.Turns,
		Data:   get,
		Thread: p.Threads,
	}
	fmt.Println("dialed broker at:", BrokerAddr)

	//rpc here is use to call broker,let broker do task

	//tcp is use for long communication, send event back....

	//run the main logic. input the IO turns
	accept := stubs.BrokerInit{}
	client.Go(stubs.Run, args, &accept, nil)

	conn, err := net.Dial("tcp", "52.201.7.150:8032")
	log.Println("distributor connected to broker stream at:", conn.RemoteAddr())
	if err != nil {
		log.Printf("failed to dial broker stream: %v", err)
		return
	}

	defer conn.Close()
	go func() {
		decode := gob.NewDecoder(conn)
		for {

			//receive event
			var msg stubs.Instruct
			errss := decode.Decode(&msg)
			if msg.Key != 0 {
				log.Printf("distributor receive key %c\n", msg.Key)
			}
			if errss != nil {
				fmt.Println("connection closed, decode error:", errss)
				close(msgs)
				return
			}
			msgs <- msg

		}
	}()

	//initflip just to count how many cell is there
	initFlips := make([]util.Cell, 0)
	for i := 0; i < y; i++ {
		for j := 0; j < x; j++ {
			if Before[i][j] == 255 {
				initFlips = append(initFlips, util.Cell{X: j, Y: i})
			}
		}
	}
	c.events <- CellsFlipped{CompletedTurns: 0, Cells: initFlips}

	turn := 0
	c.events <- StateChange{turn, Executing}

	// send keypresses

	//go function get the key form io for each key
	go func(key <-chan rune, client *rpc.Client) {
		void := stubs.Void{}
		for k := range key {
			log.Printf("Handlekey : %c\n", k)
			client.Go(stubs.Handlekey, k, &void, nil)
		}
	}(c.keyPresses, client)

	save := func(completed int, Slice [][]uint8) {
		if completed < 0 {
			completed = 0
		}
		c.ioCommand <- ioOutput
		c.ioFilename <- fmt.Sprintf("%dx%dx%d", x, y, completed)
		out := make([]uint8, x*y)
		for i := 0; i < y; i++ {
			copy(out[i*x:(i+1)*x], Slice[i])
		}
		c.ioOutput <- out
		c.ioCommand <- ioCheckIdle
		<-c.ioIdle
		c.events <- ImageOutputComplete{CompletedTurns: completed, Filename: fmt.Sprintf("%dx%dx%d", x, y, completed)}
	}

	fmt.Println("entering main event loop")
	paused := false
	//when paused and next instruct is not command
	queue := make([]stubs.Instruct, 0, 16)

	//if queue not empty process queue first
	for {
		var instruct stubs.Instruct
		if len(queue) > 0 {
			instruct = queue[0]
			queue = queue[1:]
		} else {
			msg := <-msgs
			instruct = msg
		}

		switch instruct.Name {
		case "COMMAND":
			switch instruct.Key {
			case 's':
				save(instruct.CompleteTurn, instruct.Slice)
				c.events <- StateChange{CompletedTurns: instruct.CompleteTurn, NewState: Executing}
				//send final state and leave
			case 'q':
				save(instruct.CompleteTurn, instruct.Slice)

				c.events <- FinalTurnComplete{CompletedTurns: instruct.CompleteTurn, Alive: instruct.LiveCount}
				c.events <- StateChange{CompletedTurns: instruct.CompleteTurn, NewState: Quitting}
				fmt.Println("distributor finish q")
				return
			case 'k':
				save(instruct.CompleteTurn, instruct.Slice)
				c.events <- FinalTurnComplete{CompletedTurns: instruct.CompleteTurn, Alive: instruct.LiveCount}
				c.events <- StateChange{CompletedTurns: instruct.CompleteTurn, NewState: Quitting}
				return
			case 'p':
				paused = true
				log.Printf("Paused at turn %d\n", instruct.CompleteTurn)
				c.events <- StateChange{CompletedTurns: instruct.CompleteTurn, NewState: Paused}
				for paused {
					next := <-msgs
					//if it is not Command put it in queue and continue to next turn for paused loop to get next command
					if next.Name != "COMMAND" {
						queue = append(queue, next)
						continue
					}
					fmt.Println("go to switch")

					switch next.Key {
					case 's':
						save(next.CompleteTurn, next.Slice)
						c.events <- StateChange{CompletedTurns: next.CompleteTurn, NewState: Paused}
					case 'q':

						save(next.CompleteTurn, next.Slice)
						c.events <- FinalTurnComplete{CompletedTurns: next.CompleteTurn, Alive: next.LiveCount}
						c.events <- StateChange{CompletedTurns: next.CompleteTurn, NewState: Quitting}
						return
					case 'k':
						save(next.CompleteTurn, next.Slice)
						c.events <- FinalTurnComplete{CompletedTurns: next.CompleteTurn, Alive: next.LiveCount}
						c.events <- StateChange{CompletedTurns: next.CompleteTurn, NewState: Quitting}
						client.Close()
						return
					case 'p':
						c.events <- StateChange{CompletedTurns: next.CompleteTurn, NewState: Executing}
						paused = false
					}
				}
			}

		case "TURNEND":
			c.events <- CellsFlipped{CompletedTurns: instruct.CompleteTurn, Cells: instruct.Cells}
			c.events <- TurnComplete{CompletedTurns: instruct.CompleteTurn}
		case "ALIVE":
			c.events <- AliveCellsCount{CompletedTurns: instruct.CompleteTurn, CellsCount: instruct.CellCount}
		case "FINAL":
			c.ioCommand <- ioOutput
			c.ioFilename <- fmt.Sprintf("%dx%dx%d", x, y, p.Turns)
			c.ioOutput <- instruct.Out
			c.ioCommand <- ioCheckIdle
			<-c.ioIdle
			c.events <- FinalTurnComplete{CompletedTurns: instruct.CompleteTurn, Alive: instruct.LiveCount}
			c.events <- StateChange{CompletedTurns: instruct.CompleteTurn, NewState: Quitting}
			close(c.events)
			return
		}
	}
}
