package tests

import (
	"fmt"
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

const benchLength =1000


// BenchmarkGol runs the Game of Life simulation with different thread counts.
func BenchmarkGol(b *testing.B) {
	for threads := 1; threads <= 16; threads++ {
		// disable normal program output so benchmark results stay clean
		os.Stdout = nil

		p := gol.Params{
			Turns:       benchLength,
			Threads:     threads,
			ImageWidth:  512,
			ImageHeight: 512,
		}

		name := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)

		// each sub-benchmark will be run with b.N iterations
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				events := make(chan gol.Event)
				go gol.Run(p, events, nil)
				for range events {
					// just drain the events until Run completes
				}
			}
		})
	}
}
