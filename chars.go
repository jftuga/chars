/*
chars
-John Taylor
Jan-16-2022

Determine the end-of-line format, tabs, bom, and nul
Pass wildcard filename globs on the command line
*/

package chars

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/jftuga/ellipsis"
	"github.com/olekukonko/tablewriter"
)

const pgmName string = "chars"
const pgmDesc string = "Determine the end-of-line format, tabs, bom, and nul"
const pgmUrl string = "https://github.com/jftuga/chars"
const pgmVersion string = "1.3.0"

type SpecialChars struct {
	Filename string `json:"filename"`
	Crlf     uint64 `json:"crlf"`
	Lf       uint64 `json:"lf"`
	Tab      uint64 `json:"tab"`
	Bom8     uint64 `json:"bom8"`
	Bom16    uint64 `json:"bom16"`
	Nul      uint64 `json:"nul"`
}

// Usage - display help when no cmd-line args given
func Usage() {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "%s v%s\n", pgmName, pgmVersion)
	fmt.Fprintln(os.Stderr, pgmDesc)
	fmt.Fprintln(os.Stderr, pgmUrl)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "%s [file-glob 1] [file-glob 2] ...\n", pgmName)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintln(os.Stderr, "Notes:")
	fmt.Fprintln(os.Stderr, "Use - to read filenames from STDIN")
	fmt.Fprintf(os.Stderr, "Also try: %s *  -or-  %s */*  -or-  %s */*/*\n", pgmName, pgmName, pgmName)
	fmt.Fprintf(os.Stderr, "\n")
}

// isText - if 2% of the bytes the first 1024 bytes are non-printable,
// then consider the file to be a binary file
func isText(s []byte, n int) bool {
	const max = 1024
	if n < max {
		s = s[0:n]
	}

	bin := 0
	for i, c := range string(s) {
		if i <= 3 {
			continue // skip bom characters
		}
		if i+utf8.UTFMax > len(s) {
			break // last char may be incomplete - ignore
		}
		if c == 0xFFFD || c < ' ' && c != '\n' && c != '\t' && c != '\f' && c != '\r' && c != 0x00 {
			bin += 1
		}
	}
	amount := float32(bin) / float32(len(s))
	if amount >= 0.02 {
		return false
	}
	return true
}

// detectBOM - determine byte order mark (if any)
func detectBOM(data []byte) (uint64, uint64) {
	var bom8, bom16 uint64

	if len(data) >= 3 {
		if data[0] == 239 && data[1] == 187 && data[2] == 191 {
			bom8 += 1
		}
	} else if len(data) >= 2 {
		if data[0] == 255 && data[1] == 254 {
			bom16 += 1 // little-endian
		} else if data[0] == 254 && data[1] == 255 {
			bom16 += 1 // big-endian
		}
	}

	return bom8, bom16
}

// detect - tabulate the total number of special characters
// in the given []byte slice
func detect(filename string, examineBinary bool) (SpecialChars, error) {
	var tab, lf, crlf, nul uint64
	var prev byte

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return SpecialChars{}, fmt.Errorf("warning: unable to open file: %s", filename)
	}
	defer file.Close()

	data := make([]byte, 1024)
	bytesRead, err := file.Read(data)
	if err != nil {
		return SpecialChars{}, fmt.Errorf("%s\nunable to read beginning of file: %s", err, filename)
	}

	// check binary files?
	if !examineBinary {
		if !isText(data, bytesRead) {
			return SpecialChars{}, fmt.Errorf("skipping unwanted binary file: %s", filename)
		}
	}

	// file starts with BOM?
	bom8, bom16 := detectBOM(data)

	// reset stream pointer to beginning of file
	offset, err := file.Seek(0, 0)
	if offset != 0 || err != nil {
		return SpecialChars{}, fmt.Errorf("%s\nunable to seek on file: %s", err, filename)
	}

	// incrementally read file in chunks as to not consume too much memory; search for special chars
	buf := bufio.NewReader(file)
	for {
		b, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return SpecialChars{}, err
		}

		switch b {
		case 0:
			nul++
		case '\t':
			tab++
		case '\n':
			if prev == '\r' {
				crlf++
			} else {
				lf++
			}
		}

		prev = b
	}

	return SpecialChars{Filename: filename, Crlf: crlf, Lf: lf, Tab: tab, Bom8: bom8, Bom16: bom16, Nul: nul}, nil
}

// OutputTextTable - display a text table with each filename and the number of special characters
func OutputTextTable(allStats []SpecialChars, maxLength int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"filename", "crlf", "lf", "tab", "nul", "bom8", "bom16"})

	var name string
	for _, s := range allStats {
		if maxLength == 0 {
			name = s.Filename
		} else {
			name = ellipsis.Shorten(s.Filename, maxLength)
		}
		row := []string{name, strconv.FormatUint(s.Crlf, 10), strconv.FormatUint(s.Lf, 10), strconv.FormatUint(s.Tab, 10), strconv.FormatUint(s.Nul, 10), strconv.FormatUint(s.Bom8, 10), strconv.FormatUint(s.Bom16, 10)}
		table.Append(row)
	}
	table.Render()
}

// GetJSON - return results JSON format
func GetJSON(allStats []SpecialChars) string {
	j, err := json.MarshalIndent(allStats, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	return string(j)
}

// ProcessGlob - process of files contained within a single cmd-line arg glob
func ProcessGlob(globArg string, allStats *[]SpecialChars, examineBinary bool, excludeMatched *regexp.Regexp) {
	globFiles, err1 := filepath.Glob(globArg)
	if err1 != nil {
		panic(err1)
	}
	ProcessFileList(globFiles, allStats, examineBinary, excludeMatched)
}

// ProcessFileList - process a list of filenames
func ProcessFileList(globFiles []string, allStats *[]SpecialChars, examineBinary bool, excludeMatched *regexp.Regexp) {
	for _, filename := range globFiles {
		info, _ := os.Stat(filename)
		if info.IsDir() {
			// fmt.Println("skipping directory:", filename)
			continue
		}
		if excludeMatched != nil {
			if excludeMatched.Match([]byte(filename)) {
				// fmt.Println("excluding file:", filename)
				continue
			}
		}

		// fmt.Println("checking file:", filename)
		stats, err := detect(filename, examineBinary)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}
		*allStats = append(*allStats, stats)
	}
}
