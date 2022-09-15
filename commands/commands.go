package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

func NewCmd(ctx context.Context, opt *getoptions.GetOpt, args []string) (err error) {
	cfg, err := ReadConfig()
	if err != nil {
		return
	}

	index, err := currentADRIndex(cfg.AbsDir())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return
	}

	title := strings.Join(args, " ")
	filename := fmt.Sprintf("%04d-%s.md", index+1, normalizeText(title))
	filepath := path.Join(cfg.AbsDir(), filename)
	err = newADRFile(title, filepath)
	if err != nil {
		return
	}

	fmt.Println("Created ADR:", filepath)
	return LaunchEditor(filepath)
}

func InitCmd(ctx context.Context, opt *getoptions.GetOpt, args []string) (err error) {
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
