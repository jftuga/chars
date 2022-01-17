/*
chars
-John Taylor
Jan-16-2022

Determine the end-of-line format, tabs, bom, and nul
Pass wildcard filename globs on the command line
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
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
const pgmVersion string = "1.2.0"

type FileStat struct {
	filename string
	crlf     int
	lf       int
	tab      int
	bom8     int
	bom16    int
	nul      int
}

// usage - display help when no cmd-line args given
func usage() {
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

// getFile - return the contents of a file in a []byte slice
func getFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	buf := bufio.NewReader(file)
	_, err = buf.Read(bytes)

	return bytes, err
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

	return FileStat{filename: filename, crlf: crlf, lf: lf, tab: tab, bom8: bom8, bom16: bom16, nul: nul}
}

// output - display a text table with each filename and the number of special characters
func output(allStats []FileStat, maxLength int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"filename", "crlf", "lf", "tab", "nul", "bom8", "bom16"})

	var name string
	for _, s := range allStats {
		if maxLength == 0 {
			name = s.filename
		} else {
			name = ellipsis.Shorten(s.filename, maxLength)
		}
		row := []string{name, strconv.Itoa(s.crlf), strconv.Itoa(s.lf), strconv.Itoa(s.tab), strconv.Itoa(s.nul), strconv.Itoa(s.bom8), strconv.Itoa(s.bom16)}
		table.Append(row)
	}
	table.Render()
}

// processGlob - process of files contained within a single cmd-line arg glob
func processGlob(globArg string, allStats *[]FileStat, examineBinary bool, excludeMatched *regexp.Regexp) {
	globFiles, err1 := filepath.Glob(globArg)
	if err1 != nil {
		panic(err1)
	}
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
		data, err2 := getFile(filename)
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

func main() {
	argsBinary := flag.Bool("b", false, "examine binary files")
	argsExclude := flag.String("e", "", "exclude based on regular expression; use .* instead of *")
	argsMaxLength := flag.Int("l", 0, "shorten files names to a maximum of the length")

	flag.Usage = usage
	flag.Parse()
	allGlobs := flag.Args()
	if len(allGlobs) == 0 {
		usage()
		return
	}

	var err error
	var excludeMatched *regexp.Regexp
	if len(*argsExclude) > 0 {
		excludeMatched, err = regexp.Compile(*argsExclude)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid 'exclude' regular expression: %s\n", *argsExclude)
			os.Exit(3)
		}
	}

	var allStats []FileStat
	for _, globArg := range allGlobs {
		processGlob(globArg, &allStats, *argsBinary, excludeMatched)
	}

	output(allStats, *argsMaxLength)
}
