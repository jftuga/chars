package chars

/*
case.go
-John Taylor
Jan-21-2022

An attempt to create a case-insensitive glob for file name matching on Windows which is then to be used in conjunction
with filepath.Glob().

If the name passed into CaseInsensitive contains '[' or ']', then an error will be returned.

Any PRs to improve this are most welcomed!

It does this by changing a filename string:
from: c:\bin\test.exe
to:   c:\[Bb][Ii][Nn]\[Tt][Ee][Ss][Tt].[Ee][Xx][Ee]

Caveats
-------
1) language.English is hard-coded in CaseInsensitive()
2) If a '[' or a ']' character is found in the filename, they are converted to '?' which means that filepath.Glob() may
   match more than expected.
*/

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func IsLetter(r int32) bool {
	if !unicode.IsLetter(r) {
		return false
	}
	return true
}

func CaseInsensitive(name string) string {
	if strings.Contains(name, "[") || strings.Contains(name, "]") {
		//return "", fmt.Errorf("error: input not allowed to contain '[' or ']'")
		_, _ = fmt.Fprintf(os.Stderr, "warning: filenames containing '[' or ']' can lead to unpredictable results; consider using wildcards\n")
		return name
	}
	var newName bytes.Buffer
	for i, ch := range name {
		if len(name) > 1 && i == 0 && name[i+1] == ':' {
			newName.WriteRune(ch)
		} else if ch == '[' {
			newName.WriteString(`?`)
		} else if ch == ']' {
			newName.WriteString(`?`)
		} else if IsLetter(ch) {
			u := cases.Upper(language.English).Bytes([]byte{byte(ch)})
			l := cases.Lower(language.English).Bytes([]byte{byte(ch)})
			newName.WriteString("[")
			for _, b := range u {
				newName.WriteByte(b)
			}
			for _, b := range l {
				newName.WriteByte(b)
			}
			newName.WriteString("]")
		} else {
			newName.WriteRune(ch)
		}
	}
	return newName.String()
}
