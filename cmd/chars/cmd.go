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
)

// Usage - display help when no cmd-line args given
func Usage() {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "%s v%s\n", chars.PgmName, chars.PgmVersion)
	fmt.Fprintln(os.Stderr, chars.PgmDesc)
	fmt.Fprintln(os.Stderr, chars.PgmUrl)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "%s [filename or file-glob 1] [filename or file-glob 2] ...\n", chars.PgmName)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintln(os.Stderr, "Notes:")
	fmt.Fprintln(os.Stderr, "Use - to read a file from STDIN")
	fmt.Fprintf(os.Stderr, "On Windows, try: %s *  -or-  %s */*  -or-  %s */*/*\n", chars.PgmName, chars.PgmName, chars.PgmName)
	fmt.Fprintf(os.Stderr, "\n")
}

// main - process cmd-line args; process files given on cmd-line or process file read from STDIN
func main() {
	argsBinary := flag.Bool("b", false, "examine binary files")
	argsExclude := flag.String("e", "", "exclude based on regular expression; use .* instead of *")
	argsMaxLength := flag.Int("l", 0, "shorten files names to a maximum of this length")
	argsJSON := flag.Bool("j", false, "output results in JSON format; can't be used with -l")

	flag.Usage = Usage
	flag.Parse()
	allGlobs := flag.Args()

	if *argsMaxLength > 0 && *argsJSON {
		fmt.Fprintf(os.Stderr, "-l and -j are mutually exclusive")
		os.Exit(2)
	}

	var err error
	var excludeFiles *regexp.Regexp
	if len(*argsExclude) > 0 {
		excludeFiles, err = regexp.Compile(*argsExclude)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid 'exclude' regular expression: %s\n", *argsExclude)
			os.Exit(3)
		}
	}

	// no cmd-line filenames were passed, so read from STDIN
	if len(allGlobs) == 0 {
		allGlobs = []string{"-"}
	}

	// allStats will be modified in-place by one of the two functions below
	var allStats []chars.SpecialChars
	for _, fileSelection := range allGlobs {
		if fileSelection == "-" {
			chars.ProcessStdin(&allStats, *argsBinary, excludeFiles)
		} else {
			chars.ProcessGlob(fileSelection, &allStats, *argsBinary, excludeFiles)
		}
	}

	// output results to either JSON or text table
	if *argsJSON {
		_, err := fmt.Println(chars.GetJSON(allStats))
		if err != nil {
			os.Exit(4)
		}
	} else {
		chars.OutputTextTable(allStats, *argsMaxLength)
	}
}
