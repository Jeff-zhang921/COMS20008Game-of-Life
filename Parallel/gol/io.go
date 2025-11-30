package gol

import (
	"fmt"
	"log"
	"os"
	"strconv"

	// "strings"
	"sync"

	"uk.ac.bris.cs/gameoflife/util"
)


type IOShared struct {
	command  ioCommand
	filename string
	data     []uint8
	isIdle   bool
	mu       sync.Mutex
	cond     *sync.Cond
	params   Params
}

// ioState is the internal ioState of the io goroutine.


// ioCommand allows requesting behaviour from the io (pgm) goroutine.
type ioCommand uint8

// This is a way of creating enums in Go.
// It will evaluate to:
//
//	ioOutput 	= 0
//	ioInput 	= 1
//	ioCheckIdle = 2
const (
	ioOutput ioCommand = iota
	ioInput
	ioCheckIdle
)

// writePgmImage receives an array of bytes and writes it to a pgm file.
func writePgmImage(filename string, data []uint8, width, height int) {
	_ = os.Mkdir("out", os.ModePerm)

	file, err := os.Create("out/" + filename + ".pgm")
	util.Check(err)
	defer file.Close()

	// PGM header
	_, _ = file.WriteString("P5\n")
	_, _ = file.WriteString(strconv.Itoa(width))
	_, _ = file.WriteString(" ")
	_, _ = file.WriteString(strconv.Itoa(height))
	_, _ = file.WriteString("\n")
	_, _ = file.WriteString("255\n") // maxval

	_, err = file.Write(data)
	util.Check(err)

	err = file.Sync()
	util.Check(err)

	log.Printf("[IO] File %v.pgm output done", filename)
}

// readPgmImage opens a pgm file and sends its data as an array of bytes.
func readPgmImage(filename string, width, height int) []uint8 {
	path := "../images/" + filename + ".pgm"
	f, err := os.Open(path)
	util.Check(err)
	defer f.Close()

	var magic string
	var w, h, maxval int
	_, err = fmt.Fscanf(f, "%s\n%d %d\n%d\n", &magic, &w, &h, &maxval)
	util.Check(err)
	if h != height { 
		panic(fmt.Sprintf("[IO] %v incorrect header in %s, %dx%d", util.Red("ERROR"), filename, width, height))
	}

	
	image := make([]uint8, width*height)
	n, err := f.Read(image)
	util.Check(err)
	if n != width*height {
		panic(fmt.Sprintf("[IO] %v incomplete image read in %s", util.Red("ERROR"), filename))
	}

	log.Printf("[IO] File %v.pgm input done", filename)
	return image
}



func ioWorkerShared(ioShared *IOShared) {

	for {
		ioShared.mu.Lock()
		for ioShared.isIdle { 
			ioShared.cond.Wait()
		}

		switch ioShared.command {
		case ioInput:
			fmt.Println("[IO] Input requested")
			ioShared.data = readPgmImage(ioShared.filename, ioShared.params.ImageWidth, ioShared.params.ImageHeight)
			ioShared.isIdle = true
			ioShared.cond.Signal()
		case ioOutput:
			fmt.Println("[IO] Output requested")
			writePgmImage(ioShared.filename, ioShared.data, ioShared.params.ImageWidth, ioShared.params.ImageHeight)
			ioShared.isIdle = true
			ioShared.cond.Signal()
		}

		ioShared.mu.Unlock()
	}
}
