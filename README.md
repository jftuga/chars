# chars
Determine the end-of-line format, tabs, bom, and nul characters

Binaries for Windows, macOS, Linux and FreeBSD are provided on the
[releases page](https://github.com/jftuga/chars/releases).

## Usage

* For help, run `chars -h`

```
chars v2.1.0
Determine the end-of-line format, tabs, bom, and nul
https://github.com/jftuga/chars

Usage:
chars [filename or file-glob 1] [filename or file-glob 2] ...
  -b    examine binary files
  -e string
        exclude based on regular expression; use .* instead of *
  -j    output results in JSON format; can't be used with -l
  -l int
        shorten files names to a maximum of this length

Notes:
Use - to read a file from STDIN
On Windows, try: chars *  -or-  chars */*  -or-  chars */*/*
```

## Example 1

* Run `chars` with no additional cmd-line switches
* * Only report files in the current directory
* * Report text files only since `-b` is not used

```shell
PS C:\chars> .\chars.exe *
+-----------------+------+-----+-----+------+------+-------+-----------+
|    FILENAME     | CRLF | LF  | TAB | NUL  | BOM8 | BOM16 | BYTESREAD |
+-----------------+------+-----+-----+------+------+-------+-----------+
| .goreleaser.yml |    0 |  59 |   0 |    0 |    0 |     0 |      1066 |
| LICENSE         |    0 |  21 |   0 |    0 |    0 |     0 |      1068 |
| README.md       |    0 |  92 |   0 |    0 |    0 |     0 |      3510 |
| chars.go        |    0 | 246 | 328 |    0 |    0 |     0 |      6477 |
| go.mod          |    0 |  10 |   2 |    0 |    0 |     0 |       188 |
| go.sum          |    0 |   6 |   0 |    0 |    0 |     0 |       533 |
| testfile1       |    0 |  22 |   0 | 3223 |    0 |     1 |      6448 |
+-----------------+------+-----+-----+------+------+-------+-----------+
```

## Example 2

* Run `chars` with `-e` and `-l` cmd-line switches
* * Only report files starting with `p` in the `C:\Windows\System32` directory
* * Exclude all files matching `perf.*dat`
* * Shorten filenames to a maximum length of `32`

```shell
PS C:\chars> .\chars.exe -e perf.*dat -l 32 C:\Windows\System32\p*
+----------------------------------+------+----+-----+------+------+-------+-----------+
|             FILENAME             | CRLF | LF | TAB | NUL  | BOM8 | BOM16 | BYTESREAD |
+----------------------------------+------+----+-----+------+------+-------+-----------+
| C:\Windows\System32\pcl.sep      |   11 |  0 |   0 |    0 |    0 |     0 |       150 |
| C:\Windows\System32\perfmon.msc  | 1933 |  0 |   0 |    0 |    0 |     0 |    145519 |
| C:\Windows\Sys...tmanagement.msc | 1945 |  0 |   0 |    0 |    0 |     0 |    146389 |
| C:\Windows\System32\pscript.sep  |    2 |  0 |   0 |    0 |    0 |     0 |        51 |
| C:\Windows\Sys...eryprovider.mof |    0 | 61 |   0 | 2073 |    0 |     1 |      4148 |
+----------------------------------+------+----+-----+------+------+-------+-----------+
```

## Example 3

* Pipe STDIN to `chars`
* Use JSON output, with `-j`

```shell
$ curl -s https://example.com/ | chars -j
```

```json
[
    {
        "filename": "STDIN",
        "crlf": 0,
        "lf": 46,
        "tab": 0,
        "bom8": 0,
        "bom16": 0,
        "nul": 0,
        "bytesRead": 1256
    }
]
```

## Example 4

* Fail when certain characters are detected, with `-f`
* * OS exit code on a `-f` failure is always `100`
* * `-f` is a comma-delimited list containing: `crlf`, `lf`, `tab`, `nul`, `bom8`, `bom16`

```shell
$ chars -f lf,tab /etc/group ; echo $?
+------------+------+----+-----+-----+------+-------+-----------+
|  FILENAME  | CRLF | LF | TAB | NUL | BOM8 | BOM16 | BYTESREAD |
+------------+------+----+-----+-----+------+-------+-----------+
| /etc/group |    0 | 58 |   0 |   0 |    0 |     0 |       795 |
+------------+------+----+-----+-----+------+-------+-----------+

100
```

## Reading from STDIN on Windows
* **YMMV when piping to `STDIN` under Windows**
* * Under `cmd`, instead of `type input.txt | chars`, use `<` redirection when possible: `chars < input.txt`
* * Under a recent version of `powershell`, use `Get-Content -AsByteStream input.txt | chars` instead of just `Get-Content input.txt | chars`
* `cmd` and `powershell` will skip `BOM` characters; these 2 fields will both report a value of `0`
* `cmd` and `powershell` will skip `NUL` characters; this field report a value of `0`
* `cmd` will convert `LF` to `CRLF` for `UTF-16` encoded files
* `powershell` will convert `LF` to `CRLF`
* Piping from programs such as `curl` will return `LF` characters under `cmd`, but `CRLF` under `powershell`
* * Under powershell, consider using `curl --output`

## Case Folding on Windows
* [Case folding](https://www.w3.org/TR/charmod-norm/#definitionCaseFolding) on Windows is somewhat implemented in [case.go](case.go).
* * This programs attempts case-insensitive filename matching since this is the expected behavior on Windows.
* * It is hard-coded to `English`.

## Wikipedia

* [Newline](https://en.wikipedia.org/wiki/Newline#Representation) - `CRLF` vs `LF`
* [Tab key](https://en.wikipedia.org/wiki/Tab_key#Tab_characters)
* [Null character](https://en.wikipedia.org/wiki/Null_character)
* [Byte order mark](https://en.wikipedia.org/wiki/Byte_order_mark) - `BOM-8` vs `BOM-16`

## Acknowledgments

* [ellipsis](https://github.com/jftuga/ellipsis) - Go module to insert an ellipsis into the middle of a long string to shorten it
* [tablewriter](https://github.com/olekukonko/tablewriter) - ASCII table in golang
* [/u/skeeto](https://old.reddit.com/user/skeeto) and [/u/petreus](https://old.reddit.com/user/ppetreus) provided [code review and suggestions](https://old.reddit.com/r/golang/comments/s64jye/i_wrote_a_cli_tool_to_determine_the_endofline/) 
