package main

import (
	"context"
	"fmt"

	"github.com/DavidGamba/go-getoptions"
)

func InitCommand(ctx context.Context, opt *getoptions.GetOpt, args []string) (err error) {
	dir := opt.Value("dir").(string)
	err = InitConfigFile(&Config{
		Dir: dir,
	})
	if err != nil {
		return
	}
	fmt.Printf("Initialized ADR dir to %s.\n", dir)
	return
}
