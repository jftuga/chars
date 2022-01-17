# chars
Determine the end-of-line format, tabs, bom, and nul characters

## Download

Binaries for Windows, MacOS, Linux and FreeBSD are provided on the
[releases page](https://github.com/jftuga/chars/releases).

## Usage

```shell
chars v1.2.1
Determine the end-of-line format, tabs, bom, and nul
https://github.com/jftuga/chars

Usage:
chars [file-glob 1] [file-glob 2] ...
  -b    examine binary files
  -e string
        exclude based on regular expression; use .* instead of *
  -l int
        shorten files names to a maximum of the length

Notes:
Also try: chars *  -or-  chars */*  -or-  chars */*/*
```

## Example 1

* Run `chars` with no additional cmd-line switches
* * Only report files in the current directory
* * Report text files only since `-b` is not used

```shell
PS C:\chars> .\chars.exe *

+-----------------+------+------+------+--------+------+-------+
|    FILENAME     | CRLF |  LF  | TAB  |  NUL   | BOM8 | BOM16 |
+-----------------+------+------+------+--------+------+-------+
| .gitignore      |    0 |   16 |    0 |      0 |    0 |     0 |
| .goreleaser.yml |    0 |   59 |    0 |      0 |    0 |     0 |
| LICENSE         |    0 |   21 |    0 |      0 |    0 |     0 |
| README.md       |    0 |   23 |    0 |      0 |    0 |     0 |
| cmd.go          |    0 |  148 |  148 |      0 |    0 |     0 |
| go.mod          |    0 |    7 |    0 |      0 |    0 |     0 |
| go.sum          |    0 |    4 |    0 |      0 |    0 |     0 |
| testfile1       |   19 |   19 |    0 |      0 |    0 |     0 |
| testfile2       |    0 |   21 |    0 |   2269 |    0 |     1 |
+-----------------+------+------+------+--------+------+-------+
```

## Example 2

* Run `chars` with `-e` and `-l` cmd-line switches
* * Only report files starting with `p` in the `C:\Windows\System32` directory
* * Exclude all files matching `perf.*dat`
* * Shorten filenames to a maximum length of `32`

```shell
PS C:\chars> .\chars.exe -e perf.*dat -l 32 C:\Windows\System32\p*

+----------------------------------+------+------+-----+------+------+-------+
|             FILENAME             | CRLF |  LF  | TAB | NUL  | BOM8 | BOM16 |
+----------------------------------+------+------+-----+------+------+-------+
| C:\Windows\System32\pcl.sep      |   11 |   11 |   0 |    0 |    0 |     0 |
| C:\Windows\System32\perfmon.msc  | 1933 | 1933 |   0 |    0 |    0 |     0 |
| C:\Windows\Sys...tmanagement.msc | 1945 | 1945 |   0 |    0 |    0 |     0 |
| C:\Windows\System32\pscript.sep  |    2 |    2 |   0 |    0 |    0 |     0 |
| C:\Windows\Sys...eryprovider.mof |    0 |   61 |   0 | 2073 |    0 |     1 |
+----------------------------------+------+------+-----+------+------+-------+
```

## Wikipedia

* [Newline](https://en.wikipedia.org/wiki/Newline#Representation) - `CRLF` vs `LF`
* [Tab key](https://en.wikipedia.org/wiki/Tab_key#Tab_characters)
* [Null character](https://en.wikipedia.org/wiki/Null_character)
* [Byte order mark](https://en.wikipedia.org/wiki/Byte_order_mark) - `BOM-8` vs `BOM-16`

## Acknowledgments

* [ellipsis](https://github.com/jftuga/ellipsis) - Go module to insert an ellipsis into the middle of a long string to shorten it
* [tablewriter](https://github.com/olekukonko/tablewriter) - ASCII table in golang
