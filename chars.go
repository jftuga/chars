/*
chars.go
-John Taylor
Jan-16-2022

Determine the end-of-line format, tabs, bom, and nul characters
https://github.com/jftuga/chars
*/

package chars

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jftuga/ellipsis"
	"github.com/olekukonko/tablewriter"
)

const PgmName string = "chars"
const PgmDesc string = "Determine the end-of-line format, tabs, bom, and nul"
const PgmUrl string = "https://github.com/jftuga/chars"
const PgmVersion string = "2.2.0"
const BlockSize int = 4096

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

type CharsError struct {
	code int
	err  string
}

// isText - if 2% of the bytes are non-printable, consider the file to be binary
func isText(s []byte, n int) bool {
	const binaryCutoff float32 = 0.02
	if n < BlockSize {
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

// searchForSpecialChars - search for special chars by incrementally reading in chunks as to not consume too much memory
// use *bufio.Reader so that either a file or STDIN can be used
func searchForSpecialChars(filename string, rdr *bufio.Reader, examineBinary bool) (SpecialChars, CharsError) {
	var (
		bomUtf8    = [...]byte{0xef, 0xbb, 0xbf}
		bomUtf16le = [...]byte{0xff, 0xfe}
		bomUtf16be = [...]byte{0xfe, 0xff}
	)
	var bom8, bom16 uint64
	var err error
	var bom []byte

	// check for a BOM
	// https://en.wikipedia.org/wiki/Byte_order_mark
	bom, err = rdr.Peek(2)
	if err == nil {
		if bytes.HasPrefix(bom, bomUtf16le[:2]) {
			bom16++
		} else if bytes.HasPrefix(bom[:2], bomUtf16be[:2]) {
			bom16++
		}
	}
	if bom16 == 0 {
		bom, err = rdr.Peek(3)
		if err == nil {
			if bytes.HasPrefix(bom, bomUtf8[:3]) {
				bom8++
			}
		}
	}

	// check if file contains binary data
	var firstBlock []byte
	firstBlock, err = rdr.Peek(1024)
	if err != nil {
		if err != io.EOF {
			_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}
	}
	if !examineBinary && !isText(firstBlock, 1024) {
		return SpecialChars{}, CharsError{code: 2, err: fmt.Sprintf("skipping unwanted binary file: %s", filename)}
	}

	var tab, lf, crlf, nul, bytesRead uint64

	last := byte(0)
	buff := make([]byte, 0, BlockSize)
	for {
		n, err := rdr.Read(buff[:cap(buff)])
		buff = buff[:n]
		bytesRead += uint64(n)
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			return SpecialChars{}, CharsError{code: 1, err: err.Error()}
		}

		for _, b := range buff {
			if b < ' ' {
				if b == '\x00' {
					nul++
				} else if b == '\n' {
					lf++
					if last == '\r' {
						crlf++
						lf--
					}
				} else if b == '\t' {
					tab++
				}
			}
			last = b
		}

		if err != nil && err != io.EOF {
			return SpecialChars{}, CharsError{code: 1, err: err.Error()}
		}
	}

	sc := SpecialChars{Filename: filename,
		Crlf: crlf, Lf: lf, Tab: tab, Bom8: bom8, Bom16: bom16, Nul: nul, BytesRead: bytesRead,
	}
	return sc, CharsError{code: 0, err: ""}
}

// sortByName - sorts a slice of SpecialChars by filename
/*func sortByName(entry []SpecialChars) {
	sort.Slice(entry, func(i, j int) bool {
		if strings.ToLower(entry[i].Filename) > strings.ToLower(entry[j].Filename) {
			return false
		}
		return true
	})
}*/

// OutputTextTable - display a text table with each filename and the number of special characters
func OutputTextTable(allStats []SpecialChars, maxLength int) error {
	if len(allStats) == 0 {
		return nil
	}
	// TODO: make this a cmd-line option...
	// sortByName(allStats)
	w := bufio.NewWriter(os.Stdout)
	table := tablewriter.NewWriter(w)
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
	return w.Flush()
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

// ProcessGlob - process all files matching the file-glob
func ProcessGlob(globArg string, allStats *[]SpecialChars, examineBinary bool, excludeMatched *regexp.Regexp, fail string) uint64 {
	var err error
	anyCase := CaseInsensitive(globArg)
	if len(globArg) > 0 && len(anyCase) == 0 {
		anyCase = globArg
	}

	globFiles, err := filepath.Glob(anyCase)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}
	if len(globFiles) == 0 {
		globFiles = []string{anyCase}
	}
	return ProcessFileList(globFiles, allStats, examineBinary, excludeMatched, fail)
}

// ProcessFileList - process a list of filenames
func ProcessFileList(globFiles []string, allStats *[]SpecialChars, examineBinary bool, excludeMatched *regexp.Regexp, fail string) uint64 {
	var file *os.File
	for _, filename := range globFiles {
		info, err := os.Stat(filename)
		if err != nil {
			// invalid file, skip it
			continue
		}

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
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}

		reader := bufio.NewReader(file)
		stats, charsErr := searchForSpecialChars(filename, reader, examineBinary)
		_ = file.Close()
		if charsErr.code == 1 {
			// output error message except for an unwanted binary file
			_, _ = fmt.Fprintf(os.Stderr, "error #%d: %s\n", charsErr.code, charsErr.err)
			continue
		} else if charsErr.code != 0 {
			continue
		}
		*allStats = append(*allStats, stats)
	}
	if len(fail) > 0 {
		return GetFailures(fail, allStats)
	}
	return 0
}

// ProcessStdin - read a file stream directly from STDIN
func ProcessStdin(allStats *[]SpecialChars, examineBinary bool, fail string) (uint64, CharsError) {
	var charsErr CharsError

	reader := bufio.NewReader(os.Stdin)
	stats, charsErr := searchForSpecialChars("STDIN", reader, examineBinary)
	if charsErr.code != 0 {
		return 0, charsErr
	}

	*allStats = append(*allStats, stats)
	if len(fail) > 0 {
		return GetFailures(fail, allStats), charsErr
	}
	return 0, charsErr
}

// GetFailures - parse as comma-delimited list and return the number of characters in the given character list
func GetFailures(commaList string, allStats *[]SpecialChars) uint64 {
	var failed uint64

	classes := strings.Split(strings.ToLower(commaList), ",")
	for _, class := range classes {
		for _, entry := range *allStats {
			switch class {
			case "crlf":
				failed += entry.Crlf
			case "lf":
				failed += entry.Lf
			case "tab":
				failed += entry.Tab
			case "bom8":
				failed += entry.Bom8
			case "bom16":
				failed += entry.Bom16
			case "nul":
				failed += entry.Nul
			default:
				fmt.Fprintf(os.Stderr, "Unknown character passed to -f: %s\n", class)
			}
		}
	}
	return failed
}
