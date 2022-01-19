/*
chars.go
-John Taylor
Jan-16-2022

Determine the end-of-line format, tabs, bom, and nul
Pass wildcard filename globs on the command line
*/

package chars

import (
	"bufio"
	"encoding/json"
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

const PgmName string = "chars"
const PgmDesc string = "Determine the end-of-line format, tabs, bom, and nul"
const PgmUrl string = "https://github.com/jftuga/chars"
const PgmVersion string = "2.0.0"
const firstBlockSize int = 1024

type SpecialChars struct {
	Filename  string `json:"filename"`
	Crlf      uint64 `json:"crlf"`
	Lf        uint64 `json:"lf"`
	Tab       uint64 `json:"tab"`
	Bom8      uint64 `json:"bom8"`
	Bom16     uint64 `json:"bom16"`
	Nul       uint64 `json:"nul"`
	BytesRead uint64 `json:"bytesRead"`
}

type charsError struct {
	code int
	err  string
}

// isText - if 2% of the bytes are non-printable, consider the file to be binary
func isText(s []byte, n int) bool {
	const binaryCutoff float32 = 0.02
	if n < firstBlockSize {
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
	if amount >= binaryCutoff {
		return false
	}
	return true
}

// DetectBOM - determine byte order mark (if any)
// reference: https://en.wikipedia.org/wiki/Byte_order_mark
func DetectBOM(data []byte) (uint64, uint64) {
	var bom8, bom16 uint64

	if len(data) >= 3 {
		if data[0] == 239 && data[1] == 187 && data[2] == 191 {
			bom8 += 1
		} else if len(data) >= 2 {
			if data[0] == 255 && data[1] == 254 {
				bom16 += 1 // little-endian
			} else if data[0] == 254 && data[1] == 255 {
				bom16 += 1 // big-endian
			}
		}
	}
	return bom8, bom16
}

// searchForSpecialChars - search for special chars by incrementally reading in chunks as to not consume too much memory
// use *bufio.Reader so that either a file or STDIN can be used
func searchForSpecialChars(filename string, buf *bufio.Reader, examineBinary bool) (SpecialChars, charsError) {
	var tab, lf, crlf, bom8, bom16, nul, bytesRead uint64
	var prev byte
	var firstBlock []byte

	i := 0
	verifiedTextFile := false
	for {
		b, err := buf.ReadByte()

		// return on end of file or if an i/o error occurs
		if err != nil {
			if err == io.EOF {

				return SpecialChars{Filename: filename, Crlf: crlf, Lf: lf, Tab: tab, Bom8: bom8, Bom16: bom16, Nul: nul, BytesRead: bytesRead}, charsError{code: 0, err: ""}
			}
			return SpecialChars{}, charsError{code: 1, err: err.Error()}
		}

		// create a small block at beginning of stream so that it can be checked for a BOM or contains binary data
		if i < firstBlockSize {
			firstBlock = append(firstBlock, b)
			i++
		} else {
			if !verifiedTextFile {
				verifiedTextFile = true
				bom8, bom16 = DetectBOM(firstBlock)
				if !isText(firstBlock, i) && !examineBinary {
					return SpecialChars{}, charsError{code: 2, err: fmt.Sprintf("skipping unwanted binary file: %s", filename)}
				}
			}
		}

		// iterate through each character, search for nul, tab, crlf and lf
		bytesRead++
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
	//bom8, bom16 := DetectBOM(firstBlock)
	return SpecialChars{Filename: filename, Crlf: crlf, Lf: lf, Tab: tab, Bom8: bom8, Bom16: bom16, Nul: nul, BytesRead: bytesRead}, charsError{code: 0, err: ""}
}

// OutputTextTable - display a text table with each filename and the number of special characters
func OutputTextTable(allStats []SpecialChars, maxLength int) {
	if len(allStats) == 0 {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"filename", "crlf", "lf", "tab", "nul", "bom8", "bom16", "bytesRead"})

	var name string
	for _, s := range allStats {
		if maxLength == 0 {
			name = s.Filename
		} else {
			name = ellipsis.Shorten(s.Filename, maxLength)
		}
		row := []string{name, strconv.FormatUint(s.Crlf, 10), strconv.FormatUint(s.Lf, 10),
			strconv.FormatUint(s.Tab, 10), strconv.FormatUint(s.Nul, 10), strconv.FormatUint(s.Bom8, 10),
			strconv.FormatUint(s.Bom16, 10), strconv.FormatUint(s.BytesRead, 10)}
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

	if string(j) == "null" {
		return ""
	}

	return string(j)
}

// ProcessGlob - process of files contained within a single cmd-line arg glob
func ProcessGlob(globArg string, allStats *[]SpecialChars, examineBinary bool, excludeMatched *regexp.Regexp) {
	globFiles, err := filepath.Glob(globArg)
	if err != nil {
		panic(err)
	}
	ProcessFileList(globFiles, allStats, examineBinary, excludeMatched)
}

// ProcessFileList - process a list of filenames
func ProcessFileList(globFiles []string, allStats *[]SpecialChars, examineBinary bool, excludeMatched *regexp.Regexp) {
	var file *os.File
	var err error
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
		file, err = os.Open(filename)
		if err != nil {
			fmt.Println(err)
			continue
		}

		reader := bufio.NewReader(file)
		stats, charsErr := searchForSpecialChars(filename, reader, examineBinary)
		file.Close()
		if charsErr.code == 1 {
			// output error message except for an unwanted binary file
			fmt.Fprintf(os.Stderr, "error #%d: %s\n", charsErr.code, charsErr.err)
			continue
		} else if charsErr.code != 0 {
			continue
		}
		*allStats = append(*allStats, stats)
	}
}

// ProcessStdin - process a list of filenames
func ProcessStdin(allStats *[]SpecialChars, examineBinary bool, excludeMatched *regexp.Regexp) charsError {
	var charsErr charsError

	reader := bufio.NewReader(os.Stdin)
	stats, charsErr := searchForSpecialChars("STDIN", reader, examineBinary)
	if charsErr.code != 0 {
		return charsErr
	}

	*allStats = append(*allStats, stats)
	return charsErr
}
