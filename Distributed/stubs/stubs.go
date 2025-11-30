package stubs

import (

	"uk.ac.bris.cs/gameoflife/util"
)

var Handlekey string = "Broker.Handlekey"
var Run string = "Broker.Run"
var Work string = "Workers.Working"

var Register string = "Broker.Register"

type RegisterRequest struct {
    Addr string
}


type Void struct{
}
type BrokerInit struct {

}

type BrokerInput struct {
	Width      int
	Height     int
	Turns      int
	Thread     int
	Data       []uint8
}

type Instruct struct {
	Name         string
	Key          rune
	Flip         util.Cell
	CompleteTurn int
	CellCount    int
	Filename     string
	NewState     int
	Slice        [][]uint8
	Cells        []util.Cell
	LiveCount    []util.Cell
	Out          []uint8
}

//this is for broker and worker
type WorkerResult struct {
	Changes []util.Cell

}

type Input struct {
	Key       rune 
	Start     int
	End       int
	Thislists []util.Cell
	Before    [][]uint8
	Whichturn int
	Thread int

}