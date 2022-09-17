package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/DavidGamba/go-getoptions"
	"github.com/xenoryt/go-adr/commands"
)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()

	opt.NewCommand("init", "Initializes ADR directory").SetCommandFn(commands.InitCmd).
		String("dir", "docs/adr")
	opt.NewCommand("new", "Create new ADR file.").SetCommandFn(commands.NewCmd).
		HelpSynopsisArgs("<title>")
	opt.NewCommand("update", "Updates ADR files.")
	opt.NewCommand("list", "Lists ADRs and their status.").SetCommandFn(commands.ListCmd)
	opt.NewCommand("search", "Search for ADR.")

	opt.HelpCommand("help", opt.Alias("h"))

	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Fprint(os.Stderr, opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		fmt.Fprint(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		return 1
	}

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if errors.Is(err, getoptions.ErrorHelpCalled) {
		return 1
	}
	if errors.Is(err, getoptions.ErrorParsing) {
		fmt.Fprint(os.Stderr, opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
