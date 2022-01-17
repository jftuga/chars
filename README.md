# chars
Determine the end-of-line format, tabs, bom, and nul characters
* Pass (multiple) filename wildcards on the command line
* use `-b` to also examine binary files

## Download

Binaries for Windows, MacOS, Linux and FreeBSD are provided on the
[releases page](https://github.com/jftuga/chars/releases).

## Usage

```shell
chars v1.1.0
Determine the end-of-line format, tabs, bom, and nul
https://github.com/jftuga/chars

Usage:
chars [file-glob 1] [file-glob 2] ...
  -b    examine binary files

Notes:
Also try: chars *  -or-  chars */*  -or-  chars */*/*
```

## Example

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
