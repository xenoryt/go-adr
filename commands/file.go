package commands

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/adr.tmpl
var adrTemplate string

// List of default editors from highest priority to lowest.
// Will use the first one that exists.
var DEFAULT_EDITORS = []string{
	"nvim",
	"vim",
	"pico",
	"nano",
}

var ErrInvalidFileFormat = errors.New("Invalid ADR file format")

// normalizeText converts plaintext into a normalized snake-case format without any special characters.
// Example: "Use *normalized* filenames!" => "use-normalized-filenames"
func normalizeText(text string) string {
	// Strip any symbols before and after the text
	text = regexp.MustCompile("^\\W*|\\W*$").ReplaceAllLiteralString(text, "")
	// Then replace all symbols between words with dashes
	text = regexp.MustCompile("\\W+\\b|\\b\\W+").ReplaceAllLiteralString(text, "-")
	return strings.ToLower(text)
}

func isADRFile(filename string) bool {
	indexRegex := regexp.MustCompile("^\\d{4}")

	s := indexRegex.FindString(filename)
	return s != "" && filename[len(filename)-3:] == ".md"
}

// currentADRIndex gets the current ADR file index.
func currentADRIndex(dir string) (index int, err error) {
	st, err := os.Stat(dir)
	if err != nil {
		return
	}
	if !st.IsDir() {
		return index, fmt.Errorf("Cannot read dir: is not a directory")
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	indexRegex := regexp.MustCompile("^\\d{4}")

	// Loop through the files backwards because the files are sorted by name
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].IsDir() {
			continue
		}
		s := indexRegex.FindString(files[i].Name())
		if s == "" {
			continue
		}
		return strconv.Atoi(s)
	}
	return
}

func newADRFile(title, dir string, index int) error {
	tmpl, err := template.New("adr").Parse(adrTemplate)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%04d-%s.md", index, normalizeText(title))
	filepath := path.Join(dir, filename)

	data := map[string]interface{}{
		"Index": index,
		"Title": title,
		"Date":  time.Now().Format("2006-01-02"),
	}

	buf := &bytes.Buffer{}

	err = tmpl.Execute(buf, data)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(filepath), 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, buf.Bytes(), 0644)
}

// LaunchEditor opens the file in an editor. Will use editor defined by EDITOR env var if it exists.
// Otherwise an editor will be chosen from the DEFAULT_EDITORS list.
func LaunchEditor(file string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		for _, ed := range DEFAULT_EDITORS {
			if path, _ := exec.LookPath(ed); path != "" {
				editor = ed
			}
		}
	}
	cmd := exec.Command(editor, file)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

type ADRInfo struct {
	Title   string
	Status  []*ADRStatus
	Index   int
	Created string
}

type ADRStatus struct {
	Date   string
	Status string
	Link   string
}

func ReadInfo(filepath string) (info *ADRInfo, err error) {
	//contents, err := os.ReadFile(filepath)
	reader, err := os.Open(filepath)
	if err != nil {
		return
	}
	blankLineRegex := regexp.MustCompile(`^\s*$`)
	titleRegex := regexp.MustCompile(`^#\s*(?P<index>\d+)\.\s+(?P<title>\S.+)`)
	dateRegex := regexp.MustCompile(`^([Dd]ate:?\s+)?(?P<date>\d+-\d+-\d+)(\s+(?P<status>\w+)\s*(?P<link>\[.*\]\(.*\))?)?`)

	type parser func(string, *ADRInfo) error

	parseTitle := func(line string, info *ADRInfo) error {
		matches := titleRegex.FindStringSubmatch(line)
		subnames := titleRegex.SubexpNames()

		for i, name := range subnames {
			if name == "" {
				continue
			}
			if i >= len(matches) {
				return fmt.Errorf("%w: missing %s", ErrInvalidFileFormat, name)
			}
			switch name {
			case "index":
				info.Index, _ = strconv.Atoi(matches[i])
			case "title":
				info.Title = matches[i]
			}
		}

		if len(matches) != 3 {
			return ErrInvalidFileFormat
		}

		return nil
	}

	parseCreateDate := func(line string, info *ADRInfo) error {
		dateIndex := dateRegex.SubexpIndex("date")
		matches := dateRegex.FindStringSubmatch(line)
		if dateIndex >= len(matches) {
			return ErrInvalidFileFormat
		}
		info.Created = matches[dateIndex]
		return nil
	}

	parseStatusHeader := func(line string, info *ADRInfo) error {
		if line[:9] != "## Status" {
			return ErrInvalidFileFormat
		}
		return nil
	}

	parseStatus := func(line string, info *ADRInfo) error {
		dateIndex := dateRegex.SubexpIndex("date")
		statusIndex := dateRegex.SubexpIndex("status")
		linkIndex := dateRegex.SubexpIndex("link")
		matches := dateRegex.FindStringSubmatch(line)
		if statusIndex >= len(matches) {
			return ErrInvalidFileFormat
		}
		status := &ADRStatus{
			Date:   matches[dateIndex],
			Status: matches[statusIndex],
		}
		if linkIndex < len(matches) {
			status.Link = matches[linkIndex]
		}
		info.Status = append(info.Status, status)
		return nil
	}

	parsers := []parser{
		parseTitle,
		parseCreateDate,
		parseStatusHeader,
	}

	info = &ADRInfo{}
	pInd := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if blankLineRegex.MatchString(line) {
			continue
		}
		if pInd < len(parsers) {
			if err = parsers[pInd](line, info); err != nil {
				return info, err
			}
			pInd++
		} else {
			if err = parseStatus(line, info); err != nil {
				return info, nil
			}
		}
	}

	return info, err
}

func ADRFiles(cfg *Config) ([]string, error) {
	dir := cfg.AbsDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	results := []string{}
	for _, entry := range files {
		if isADRFile(entry.Name()) {
			filepath := path.Join(dir, entry.Name())
			results = append(results, filepath)
		}
	}

	return results, nil
}
