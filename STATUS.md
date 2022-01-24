**Discussion**

* https://old.reddit.com/r/golang/comments/s64jye/i_wrote_a_cli_tool_to_determine_the_endofline/

**Done**

* **2022-01-21:**
* With files named: `a*b`and `aab`, `chars a*b` now works correctly
* *Case folding* for Windows only is somewhat implemented in [case.go](case.go)
* * This can be disabled by the `-s` switch; thus allowing for character classes in between a set of `[]`
* * PRs to improve this are most welcomed
* `chars.OutputTextTable()` now returns an error when output can not be displayed:
* * Example: `chars /dev/null > /dev/full`
* Added to `README.md`:
* * Use`<` redirection when possible instead of `|`
* * Using `Get-Content -AsByteStream` with newer versions of powershell
* * Under powershell, consider using `curl --output`
* When reading file chunks, the block size has been changed from `1024` to `4096` -- a small performance improvement
* `chars.searchForSpecialChars()` has been rewritten to be significantly more performant.

* **2022-01-19:**
* `chars` is now a package, with a corresponding `cmd/chars/cmd.go` for the CLI interface
* This will now report special chars for STDIN
* * cat filename | chars -
* Running `chars` by itself with no cmd-line args will also read from STDIN
* A `BytesRead` column was added to the output
* * Depending on how data is piped into `chars` *(and which OS)*, values can differ vs reading the same filename from the cmd-line. *See below*
* If a glob doesn't match any files, no output or error will occur
* Added `-j` output to create JSON format
* I think globbing is fixed for cross-platform.  
* * I am still globbing because I want this feature for Windows, but since Linux bash, macOS zsh already expand file names beforehand this should not cause a problem.
* * The program is able to handle files on Linux actually containing a `*` in their name. 
* * Please let me know if you can think of any problems that may arise with this method.
* It reads files in chunks as to not consume too much memory.

**To Do**
* ~~On Windows, filenames are case-sensitive, but expected behavior is case-insensitive~~
* ~~Checking for errors when writing output has only been implemented for json output and not for `table.Render()`~~
* ~~I am not very happy with `searchForSpecialChars()`.  It works, but seems kludgy -- how could this be function be simplified?~~
* ~~You mentioned *"paths are a tricky business"*. I'm OK if my program does not work for unconventional, non-standard file names.~~
* Create `chars_test.go` once everything has been stabilized.
* Create API examples in `README.md`

**Reading from STDIN on Windows**
* **YMMV when piping to `STDIN` under Windows**
* `cmd` and `powershell` will skip `BOM` characters; these 2 fields will both report a value of `0`
* `cmd` and `powershell` will skip `NUL` characters; this field report a value of `0`
* `cmd` will convert `LF` to `CRLF` for `UTF-16` encoded files
* `powershell` will convert `LF` to `CRLF`
* Piping from programs such as `curl` will return `LF` characters under `cmd`, but `CRLF` under `powershell`
