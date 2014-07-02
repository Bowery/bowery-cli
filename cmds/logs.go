// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"os"
	"os/signal"
	"path/filepath"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/gopackages/keen"

	"github.com/ActiveState/tail"
)

func init() {
	Cmds["logs"] = &Cmd{logsRun, "logs", "Tail your application's logs.", ""}
}

func logsRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	// Create and register signals.
	signals := make(chan os.Signal, 1)
	done := make(chan int, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	defer signal.Stop(signals)

	dev, err := db.GetDeveloper()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	keen.AddEvent("bowery logs", map[string]*db.Developer{"user": dev})

	db.GetState() // Do this to ensure we're at the root of the app.
	tailer, err := tail.TailFile(filepath.Join(".bowery", "output.log"), tail.Config{
		MustExist: true,
		Follow:    true,
		Logger:    tail.DiscardingLogger,
	})
	if err != nil {
		if !os.IsNotExist(err) {
			err = errors.NewStackError(err)
		} else {
			err = errors.ErrNotConnected
		}

		rollbar.Report(err)
		return 1
	}
	defer tailer.Stop()

	// Catch signals and finish.
	go func() {
		<-signals
		done <- 0
	}()

	// Print lines as they come in, on error finish.
	go func() {
		for line := range tailer.Lines {
			if line.Err != nil {
				rollbar.Report(errors.NewStackError(line.Err))
				done <- 1
				return
			}

			log.Println("", line.Text)
		}
	}()

	return <-done
}
