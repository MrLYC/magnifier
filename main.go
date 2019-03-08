package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/mrlyc/magnifier/logging"
	"github.com/mrlyc/magnifier/magnifier"
)

type initialHandler func() bool

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&magnifier.Command{}, "")

	level := "info"
	flag.StringVar(&level, "log-level", level, "logger level")
	flag.Parse()

	logger := logging.GetLogger()
	err := logging.SetLevel(logger, level)
	if err != nil {
		panic(err)
	}
	logger.Out = os.Stderr

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
