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
	"github.com/jftuga/chars"
	"os"
	"regexp"
)

func getFileListFromStdIn() []string {
	var input *bufio.Scanner
	input = bufio.NewScanner(os.Stdin)

	var allFilenames []string
	for input.Scan() {
		allFilenames = append(allFilenames, input.Text())
	}
	return allFilenames
}

func main() {
	argsBinary := flag.Bool("b", false, "examine binary files")
	argsExclude := flag.String("e", "", "exclude based on regular expression; use .* instead of *")
	argsMaxLength := flag.Int("l", 0, "shorten files names to a maximum of the length")
	argsJSON := flag.Bool("j", false, "output results in JSON format")

	flag.Usage = chars.Usage
	flag.Parse()
	allGlobs := flag.Args()
	if len(allGlobs) == 0 {
		chars.Usage()
		os.Exit(1)
	}

	if *argsMaxLength > 0 && *argsJSON {
		fmt.Fprintf(os.Stderr, "-l and -j are mutually exclusive")
		os.Exit(2)
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

	var allStats []chars.FileStat
	for _, globArg := range allGlobs {
		if globArg == "-" {
			allFiles := getFileListFromStdIn()
			chars.ProcessFileList(allFiles, &allStats, *argsBinary, excludeMatched)
		} else {
			chars.ProcessGlob(globArg, &allStats, *argsBinary, excludeMatched)
		}
	}

	if *argsJSON {
		fmt.Println(chars.GetJSON(allStats))
	} else {
		chars.OutputTextTable(allStats, *argsMaxLength)
	}
}
