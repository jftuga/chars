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
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jftuga/ellipsis"
	"github.com/olekukonko/tablewriter"
)

const PgmName string = "chars"
const PgmDesc string = "Determine the end-of-line format, tabs, bom, nul and non-ascii"
const PgmUrl string = "https://github.com/jftuga/chars"
const PgmVersion string = "2.7.0"
const BlockSize int = 4096

type SpecialChars struct {
	Filename               string `json:"filename"`
	Crlf                   uint64 `json:"crlf"`
	Lf                     uint64 `json:"lf"`
	Tab                    uint64 `json:"tab"`
	Bom8                   uint64 `json:"bom8"`
	Bom16                  uint64 `json:"bom16"`
	Nul                    uint64 `json:"nul"`
	NonAscii               uint64 `json:"nonAscii"`
	MaxConsecutiveNonAscii uint64 `json:"maxConsecutiveNonAscii"`
	BytesRead              uint64 `json:"bytesRead"`
	Failure                bool   `json:"failure"`
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

	var tab, lf, crlf, nul, nonAscii, currentNonASCIIStreak, maxConsecutiveNonASCII, bytesRead uint64

	last := byte(0)
	buff := make([]byte, BlockSize)
	for {
		n, err := rdr.Read(buff)
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
			if b > 127 {
				currentNonASCIIStreak++
				if currentNonASCIIStreak > maxConsecutiveNonASCII {
					maxConsecutiveNonASCII = currentNonASCIIStreak
				}
			} else {
				currentNonASCIIStreak = 0
			}

			if b < ' ' {
				if b == 0 {
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
			} else if b > 127 {
				nonAscii++
			}
			last = b
		}

		if err != nil && err != io.EOF {
			return SpecialChars{}, CharsError{code: 1, err: err.Error()}
		}
	}

	sc := SpecialChars{Filename: filename,
		Crlf: crlf, Lf: lf, Tab: tab, Bom8: bom8, Bom16: bom16, Nul: nul, NonAscii: nonAscii,
		MaxConsecutiveNonAscii: maxConsecutiveNonASCII, BytesRead: bytesRead,
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
func OutputTextTable(allStats []SpecialChars, maxLength int, wantTotals, wantCommas bool) error {
	if len(allStats) == 0 {
		return nil
	}
	// TODO: make this a cmd-line option...
	// sortByName(allStats)
	w := bufio.NewWriter(os.Stdout)
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"filename", "crlf", "lf", "tab", "nul", "bom8", "bom16", "non-ASCII", "max consec N-A", "bytesRead"})

	var name string
	var crlf, lf, tab, nul, bom8, bom16, nonAscii, maxConsecutiveNonAscii, bytesRead uint64
	for _, s := range allStats {
		if maxLength == 0 {
			name = s.Filename
		} else {
			name = ellipsis.Shorten(s.Filename, maxLength)
		}
		row := []string{name, strconv.FormatUint(s.Crlf, 10), strconv.FormatUint(s.Lf, 10),
			strconv.FormatUint(s.Tab, 10), strconv.FormatUint(s.Nul, 10), strconv.FormatUint(s.Bom8, 10),
			strconv.FormatUint(s.Bom16, 10), strconv.FormatUint(s.NonAscii, 10),
			strconv.FormatUint(s.MaxConsecutiveNonAscii, 10), strconv.FormatUint(s.BytesRead, 10)}
		if wantCommas {
			row[1] = RenderInteger("#,###.", int64(s.Crlf))
			row[2] = RenderInteger("#,###.", int64(s.Lf))
			row[3] = RenderInteger("#,###.", int64(s.Tab))
			row[4] = RenderInteger("#,###.", int64(s.Nul))
			row[5] = RenderInteger("#,###.", int64(s.Bom8))
			row[6] = RenderInteger("#,###.", int64(s.Bom16))
			row[7] = RenderInteger("#,###.", int64(s.NonAscii))
			row[8] = RenderInteger("#,###.", int64(s.MaxConsecutiveNonAscii))
			row[9] = RenderInteger("#,###.", int64(s.BytesRead))
		}
		if wantTotals {
			crlf += s.Crlf
			lf += s.Lf
			tab += s.Tab
			nul += s.Nul
			bom8 += s.Bom8
			bom16 += s.Bom16
			nonAscii += s.NonAscii
			maxConsecutiveNonAscii += s.MaxConsecutiveNonAscii
			bytesRead += s.BytesRead
		}
		table.Append(row)
	}
	if wantTotals {
		totals := fmt.Sprintf("TOTALS: %d files", len(allStats))
		row := []string{totals, strconv.FormatUint(crlf, 10), strconv.FormatUint(lf, 10),
			strconv.FormatUint(tab, 10), strconv.FormatUint(nul, 10), strconv.FormatUint(bom8, 10),
			strconv.FormatUint(bom16, 10), strconv.FormatUint(nonAscii, 10),
			"---", strconv.FormatUint(bytesRead, 10)}
		if wantCommas {
			row[1] = RenderInteger("#,###.", int64(crlf))
			row[2] = RenderInteger("#,###.", int64(lf))
			row[3] = RenderInteger("#,###.", int64(tab))
			row[4] = RenderInteger("#,###.", int64(nul))
			row[5] = RenderInteger("#,###.", int64(bom8))
			row[6] = RenderInteger("#,###.", int64(bom16))
			row[7] = RenderInteger("#,###.", int64(nonAscii))
			row[8] = "---"
			row[9] = RenderInteger("#,###.", int64(bytesRead))
		}
		table.Append(row)
	}
	table.Render()
	return w.Flush()
}

// OutputFailedFileList - only display a list of file names that have failed when using -F cmd line option
func OutputFailedFileList(allStats []SpecialChars) {
	if len(allStats) == 0 {
		return
	}

	for _, stat := range allStats {
		if stat.Failure {
			_, _ = fmt.Println(stat.Filename)
		}
	}
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
	var failed, totalFailures uint64
	var failedAllStats []SpecialChars

	classes := strings.Split(strings.ToLower(commaList), ",")
	for i, entry := range *allStats {
		failed = 0
		for _, class := range classes {
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
			case "nonascii":
				failed += entry.NonAscii
			case "maxconsec":
				failed += entry.MaxConsecutiveNonAscii
			case "nul":
				failed += entry.Nul
			default:
				fmt.Fprintf(os.Stderr, "Unknown character passed to -f: %s\n", class)
			}
		}

		if failed > 0 {
			totalFailures += failed
			(*allStats)[i].Failure = true
			failedAllStats = append(failedAllStats, (*allStats)[i])
		}
	}
	return totalFailures
}

// sortByColumn - sorts a slice of SpecialChars by the specified column
func SortByColumn(entries []SpecialChars, column string) {
	// Define a mapping of column names to comparison functions
	compareFuncs := map[string]func(i, j int) bool{
		"filename": func(i, j int) bool {
			return strings.ToLower(entries[i].Filename) < strings.ToLower(entries[j].Filename)
		},
		"crlf": func(i, j int) bool {
			return entries[i].Crlf < entries[j].Crlf
		},
		"lf": func(i, j int) bool {
			return entries[i].Lf < entries[j].Lf
		},
		"tab": func(i, j int) bool {
			return entries[i].Tab < entries[j].Tab
		},
		"nul": func(i, j int) bool {
			return entries[i].Nul < entries[j].Nul
		},
		"bom8": func(i, j int) bool {
			return entries[i].Bom8 < entries[j].Bom8
		},
		"bom16": func(i, j int) bool {
			return entries[i].Bom16 < entries[j].Bom16
		},
		"nonascii": func(i, j int) bool {
			return entries[i].NonAscii < entries[j].NonAscii
		},
		"maxconsec": func(i, j int) bool {
			return entries[i].MaxConsecutiveNonAscii < entries[j].MaxConsecutiveNonAscii
		},
		"bytesread": func(i, j int) bool {
			return entries[i].BytesRead < entries[j].BytesRead
		},
	}

	// Get the comparison function for the specified column
	compareFunc, ok := compareFuncs[strings.ToLower(column)]
	if !ok {
		// Default to sorting by filename if the column is not recognized
		compareFunc = compareFuncs["filename"]
	}

	// Sort the entries
	sort.Slice(entries, compareFunc)
}

// GetValidSortColumns - returns a list of valid column names for sorting
func GetValidSortColumns() []string {
	return []string{
		"filename", "crlf", "lf", "tab", "nul", "bom8", "bom16", "nonascii", "maxconsec", "bytesread",
	}
}
