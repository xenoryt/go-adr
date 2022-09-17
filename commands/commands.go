package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

func NewCmd(ctx context.Context, opt *getoptions.GetOpt, args []string) (err error) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Missing title!")
		fmt.Fprintln(os.Stderr, opt.Help())
		return getoptions.ErrorHelpCalled
	}

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
	err = newADRFile(title, cfg.AbsDir(), index+1)
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

func ListCmd(ctx context.Context, opt *getoptions.GetOpt, args []string) (err error) {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	dir := cfg.AbsDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	results := make([]*ADRInfo, len(files))
	for _, entry := range files {
		if isADRFile(entry.Name()) {
			filepath := path.Join(dir, entry.Name())
			info, err := ReadInfo(filepath)
			if err != nil {
				return fmt.Errorf("Failed to parse %s: %w", filepath, err)
			}
			results = append(results, info)
			fmt.Printf("%s: %s\n", info.Title, info.Status[len(info.Status)-1].Status)
		}
	}
	return
}
