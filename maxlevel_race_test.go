package gologger

import (
	"sync"
	"testing"

	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
)

type discardWriter struct{}

func (discardWriter) Write(_ []byte, _ levels.Level) {
}

func TestSetMaxLevelConcurrent(t *testing.T) {
	l := &Logger{}
	l.SetMaxLevel(levels.LevelInfo)
	l.SetFormatter(formatter.NewCLI(true))
	l.SetWriter(discardWriter{})

	const goroutines = 8
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				l.SetMaxLevel(levels.Level(j % int(levels.LevelVerbose+1)))
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				l.Info().Msg("concurrent")
				_ = l.Enabled(t.Context(), 0)
			}
		}()
	}
	wg.Wait()
}
