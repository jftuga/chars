/*
chars
-John Taylor
Jan-16-2022

Determine the end-of-line format, tabs, bom, and nul
Pass wildcard filename globs on the command line
*/

package chars

import (
	"encoding/json"
	"flag"
	"fmt"
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

type FileStat struct {
	Filename string `json:"filename"`
	Crlf     int    `json:"crlf"`
	Lf       int    `json:"lf"`
	Tab      int    `json:"tab"`
	Bom8     int    `json:"bom8"`
	Bom16    int    `json:"bom16"`
	Nul      int    `json:"nul"`
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
func isText(s []byte) bool {
	const max = 1024 // at least utf8.UTFMax
	if len(s) > max {
		s = s[0:max]
	}

	bin := 0
	for i, c := range string(s) {
		if i <= 3 {
			// skip bom characters
			continue
		}
		if i+utf8.UTFMax > len(s) {
			// last char may be incomplete - ignore
			break
		}
		if c == 0xFFFD || c < ' ' && c != '\n' && c != '\t' && c != '\f' && c != '\r' && c != 0x00 {
			// decoding error or control character - not a text file
			bin += 1
		}
	}
	numerator := float32(bin) / float32(len(s))
	if numerator > 0.02 {
		return false
	}
	return true
}

// detect - tabulate the total number of special characters
// in the given []byte slice
func detect(filename string, data []byte) FileStat {
	position := 0 // file position
	tab := 0      // 9
	lf := 0       // 10
	crlf := 0     // 13
	bom8 := 0     // 239,187,191
	bom16 := 0    // 255,254
	nul := 0      // 0
	for i, b := range data {
		if b == 10 {
			lf += 1
		} else if b == 9 {
			tab += 1
		} else if b == 0 {
			nul += 1
		}
		if position >= 1 {
			if b == 10 && data[i-1] == 13 {
				crlf += 1
			}
		}
		position += 1
	}

	if len(data) >= 3 {
		if data[0] == 239 && data[1] == 187 && data[2] == 191 {
			bom8 += 1
		} else if data[0] == 255 && data[1] == 254 {
			bom16 += 1 // little-endian
		} else if data[0] == 254 && data[1] == 255 {
			bom16 += 1 // big-endian
		}
	}

	return FileStat{Filename: filename, Crlf: crlf, Lf: lf, Tab: tab, Bom8: bom8, Bom16: bom16, Nul: nul}
}

// OutputTextTable - display a text table with each filename and the number of special characters
func OutputTextTable(allStats []FileStat, maxLength int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"filename", "crlf", "lf", "tab", "nul", "bom8", "bom16"})

	var name string
	for _, s := range allStats {
		if maxLength == 0 {
			name = s.Filename
		} else {
			name = ellipsis.Shorten(s.Filename, maxLength)
		}
		row := []string{name, strconv.Itoa(s.Crlf), strconv.Itoa(s.Lf), strconv.Itoa(s.Tab), strconv.Itoa(s.Nul), strconv.Itoa(s.Bom8), strconv.Itoa(s.Bom16)}
		table.Append(row)
	}
	table.Render()
}

// GetJSON - return results JSON format
func GetJSON(allStats []FileStat) string {
	j, err := json.MarshalIndent(allStats, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	return string(j)
}

// ProcessGlob - process of files contained within a single cmd-line arg glob
func ProcessGlob(globArg string, allStats *[]FileStat, examineBinary bool, excludeMatched *regexp.Regexp) {
	globFiles, err1 := filepath.Glob(globArg)
	if err1 != nil {
		panic(err1)
	}
	ProcessFileList(globFiles, allStats, examineBinary, excludeMatched)
}

// ProcessFileList - process a list of filenames
func ProcessFileList(globFiles []string, allStats *[]FileStat, examineBinary bool, excludeMatched *regexp.Regexp) {
	for _, filename := range globFiles {
		info, _ := os.Stat(filename)
		if info.IsDir() {
			// fmt.Printf("skipping directory: %s\n", filename)
			continue
		}
		if excludeMatched != nil {
			if excludeMatched.Match([]byte(filename)) {
				// fmt.Printf("excluding file: %s\n", filename)
				continue
			}
		}
		data, err2 := os.ReadFile(filename)
		if !examineBinary {
			if !isText(data) {
				// fmt.Printf("skipping binary file: %s\n", filename)
				continue
			}
		}
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Unable to process: %s\n", filename)
			continue
		}

		// fmt.Println(filename)
		stats := detect(filename, data)
		*allStats = append(*allStats, stats)
	}
}
