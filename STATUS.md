**Discussion**

* https://old.reddit.com/r/golang/comments/s64jye/i_wrote_a_cli_tool_to_determine_the_endofline/

**Done**

* `chars` is now a package, with a corresponding `cmd/chars/cmd.go` for the CLI interface
* This will now report special chars for STDIN
* * cat filename | chars -
* Running `chars` by itself with no cmd-line args will also read from STDIN
* A `BytesRead` column was added to the output
* * Depending on how data is piped into `chars` *(and which OS)*, values can differ vs reading the same filename from the cmd-line. *See below*
* If a glob doesn't match any files, no output or error will occur
* Added `-j` output to create JSON format
* I think globbing is fixed for cross-platform.  
* * I am still globbing because I want this feature for Windows, but since Linux bash, MacOS zsh already expand file names beforehand this should not cause a problem.
* * The program is able to handle files on Linux actually containing a `*` in their name. 
* * Please let me know if you can think of any problems that may arise with this method.
* It reads files in chunks as to not consume too much memory.

**To Do**
* On Windows, filenames are case-sensitive, but expected behavior is case-insensitive
* * Do you think something this would be a good fix: https://wenzr.wordpress.com/2018/04/09/go-glob-case-insensitive/
* Checking for errors when writing output has only been implemented for json output and not for `table.Render()`
* * Could you please expand on: *"Use a buffered writer and, as I had just pointed out, check the result of Flush after Render returns."* as I am not familiar with this?
* I am not very happy with `searchForSpecialChars()`.  It works, but seems kludgy -- how could this be function be simplified?
* You mentioned *"paths are a tricky business"*. I'm OK if my program does not work for unconventional, non-standard file names.
* Create `chars_test.go` once everything has been stabilized.
* Create API examples in `README.md`

**Reading from STDIN on Windows**
* **YMMV when piping to `STDIN` under Windows**
* `cmd` and `powershell` will skip `BOM` characters; these 2 fields will both report a value of `0`
* `cmd` and `powershell` will skip `NUL` characters; this field report a value of `0`
* `cmd` will convert `LF` to `CRLF` for `UTF-16` encoded files
* `powershell` will convert `LF` to `CRLF`
* Piping from programs such as `curl` will return `LF` characters under `cmd`, but `CRLF` under `powershell`
