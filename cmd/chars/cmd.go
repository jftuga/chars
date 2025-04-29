/*
chars CLI interface
-John Taylor
Jan-16-2022

Determine the end-of-line format, tabs, bom, and nul characters
https://github.com/jftuga/chars
*/

package main

import (
	"flag"
	"fmt"
	"github.com/jftuga/chars"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// Usage - display help when no cmd-line args given
func Usage() {
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	_, _ = fmt.Fprintf(os.Stderr, "%s v%s\n", chars.PgmName, chars.PgmVersion)
	_, _ = fmt.Fprintln(os.Stderr, chars.PgmDesc)
	_, _ = fmt.Fprintln(os.Stderr, chars.PgmUrl)
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	_, _ = fmt.Fprintln(os.Stderr, "Usage:")
	_, _ = fmt.Fprintf(os.Stderr, "%s [filename or file-glob 1] [filename or file-glob 2] ...\n", chars.PgmName)
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	_, _ = fmt.Fprintln(os.Stderr, "Notes:")
	_, _ = fmt.Fprintln(os.Stderr, "Use - to read a file from STDIN")
	_, _ = fmt.Fprintf(os.Stderr, "On Windows, try: %s *  -or-  %s */*  -or-  %s */*/*\n", chars.PgmName, chars.PgmName, chars.PgmName)
	_, _ = fmt.Fprintf(os.Stderr, "\n")
}

// main - process cmd-line args; process files given on cmd-line or process file read from STDIN
func main() {
	argsBinary := flag.Bool("b", false, "examine binary files")
	argsExclude := flag.String("e", "", "exclude based on regular expression; use .* instead of *")
	argsMaxLength := flag.Int("l", 0, "shorten files names to a maximum of this length")
	argsJSON := flag.Bool("j", false, "output results in JSON format; can't be used with -l; does not honor -t or -c")
	argsVersion := flag.Bool("v", false, "display version and then exit")
	argsFail := flag.String("f", "", "fail with OS exit code=100 if any of the included characters exist; ex: -f crlf,nul,bom8,nonascii")
	argsFailedFileList := flag.Bool("F", false, "when used with -f, only display a list of failed files, one per line")
	argsTotals := flag.Bool("t", false, "append a row which includes a total for each column")
	argsComma := flag.Bool("c", false, "add comma thousands separator to numeric values")
	argsSortBy := flag.String("s", "filename", "sort output by column: filename, crlf, lf, tab, nul, bom8, bom16, nonascii, bytesread")

	flag.Usage = Usage
	flag.Parse()
	allGlobs := flag.Args()

	if *argsVersion {
		fmt.Printf("%s\n", chars.PgmVersion)
		os.Exit(0)
	}

	if *argsMaxLength > 0 && *argsJSON {
		_, _ = fmt.Fprintf(os.Stderr, "-l and -j are mutually exclusive")
		os.Exit(2)
	}

	// Validate sort column
	validSortColumns := chars.GetValidSortColumns()
	sortColumnValid := false
	sortColumn := strings.ToLower(*argsSortBy)
	for _, col := range validSortColumns {
		if sortColumn == col {
			sortColumnValid = true
			break
		}
	}
	if !sortColumnValid {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid sort column: %s\nValid columns are: %s\n",
			*argsSortBy, strings.Join(validSortColumns, ", "))
		os.Exit(3)
	}

	var err error
	var excludeFiles *regexp.Regexp
	if len(*argsExclude) > 0 {
		excludeFiles, err = regexp.Compile(*argsExclude)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid 'exclude' regular expression: %s\n", *argsExclude)
			os.Exit(4)
		}
	}

	// no cmd-line filenames were passed, so read from STDIN
	if len(allGlobs) == 0 {
		allGlobs = []string{"-"}
	}

	// allStats will be modified in-place by one of the two functions below
	var allStats []chars.SpecialChars
	var failed, current uint64
	for _, fileSelection := range allGlobs {
		if fileSelection == "-" {
			current, _ = chars.ProcessStdin(&allStats, *argsBinary, *argsFail)
		} else {
			if runtime.GOOS == "windows" {
				current = chars.ProcessGlob(fileSelection, &allStats, *argsBinary, excludeFiles, *argsFail)
			} else {
				current = chars.ProcessFileList([]string{fileSelection}, &allStats, *argsBinary, excludeFiles, *argsFail)
			}
		}
		failed += current
	}

	// Sort the results by the specified column
	if len(allStats) > 0 {
		chars.SortByColumn(allStats, sortColumn)
	}

	// output results to either JSON or text table
	if *argsJSON {
		_, err := fmt.Println(chars.GetJSON(allStats))
		if err != nil {
			os.Exit(5)
		}
	} else if *argsFailedFileList && len(*argsFail) > 0 && failed > 0 {
		chars.OutputFailedFileList(allStats)
	} else {
		err := chars.OutputTextTable(allStats, *argsMaxLength, *argsTotals, *argsComma)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(6)
		}
	}

	if len(*argsFail) > 0 && failed > 0 {
		os.Exit(100)
	}
}
