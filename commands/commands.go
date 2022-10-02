package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

func InitOpts(opt *getoptions.GetOpt) {
	var cmdopt *getoptions.GetOpt

	// Init Command
	cmdopt = opt.NewCommand("init", "Initializes ADR directory").SetCommandFn(InitCmd)
	cmdopt.String("dir", "docs/adr")

	// New Command
	opt.NewCommand("new", "Create new ADR file.").SetCommandFn(NewCmd).
		HelpSynopsisArgs("<title>")

		// Update Command
	opt.NewCommand("update", "Updates ADR files.")

	// List Command
	opt.NewCommand("list", "Lists ADRs and their status.").SetCommandFn(ListCmd)

	// Search Command
	cmdopt = opt.NewCommand("search", "Search for ADR with regex pattern.").SetCommandFn(SearchCmd).
		HelpSynopsisArgs("<pattern>")
	cmdopt.String("tags", "", opt.Alias("t"))
}

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
	files, err := ADRFiles(cfg)
	if err != nil {
		return
	}

	results := make([]*ADRInfo, len(files))
	for _, filepath := range files {
		info, err := ReadInfo(filepath)
		if err != nil {
			return fmt.Errorf("Failed to parse %s: %w", filepath, err)
		}
		results = append(results, info)
		fmt.Printf("%s: %s\n", info.Title, info.Status[len(info.Status)-1].Status)
	}
	return
}

func SearchCmd(ctx context.Context, opt *getoptions.GetOpt, args []string) (err error) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Missing search regex!")
		fmt.Fprintln(os.Stderr, opt.Help())
		return getoptions.ErrorHelpCalled
	}

	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	files, err := ADRFiles(cfg)
	if err != nil {
		return err
	}

	re, err := regexp.Compile(args[0])
	if err != nil {
		return err
	}

	for _, filepath := range files {
		data, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("Failed to open file: %w", err)
		}
		if re.Match(data) {
			info, err := ReadInfo(filepath)
			if err != nil {
				return fmt.Errorf("Failed to parse %s: %w", filepath, err)
			}
			fmt.Printf("%s: %s\n", info.Title, info.Status[len(info.Status)-1].Status)
		}
	}

	return
}
