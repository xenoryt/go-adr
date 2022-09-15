package commands

import (
	"bytes"
	_ "embed"
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

// normalizeText converts plaintext into a normalized snake-case format without any special characters.
// Example: "Use *normalized* filenames!" => "use-normalized-filenames"
func normalizeText(text string) string {
	// Strip any symbols before and after the text
	text = regexp.MustCompile("^\\W*|\\W*$").ReplaceAllLiteralString(text, "")
	// Then replace all symbols between words with dashes
	text = regexp.MustCompile("\\W+\\b|\\b\\W+").ReplaceAllLiteralString(text, "-")
	return strings.ToLower(text)
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

	indexRegex, _ := regexp.Compile("^\\d{4}")

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

func newADRFile(title, filepath string) error {
	tmpl, err := template.New("adr").Parse(adrTemplate)
	if err != nil {
		return err
	}

	data := map[string]string{
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
