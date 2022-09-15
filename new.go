package main

import (
	"context"
	"fmt"

	"github.com/DavidGamba/go-getoptions"
)

func NewCommand(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	fmt.Print("Args: ")
	for _, arg := range args {
		fmt.Print(arg)
	}
	return nil
}
