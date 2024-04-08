# GOSR

GOSR (Go Search and Replace) is a command line utility meant to support numerous regex-based file operations traditionally solved using some combination of `find`, `awk`, `sed`, and `grep`. This gives one consistent regex syntax while supporting many different operations.

## Commands

GOSR works using four top-level commands to solve various problem

### find

`gosr find` finds _files_ which match a given pattern, and prints them newline-delimited to stdout. `-r` is a common flag to recursively search subdirectories. It's important to note (especially when searching subdirectories) the regex is applied against the entire (relative) path, which means the regex can be used to match against the path, too, not just the filename.

```bash
$ gosr find '\.go$' -r
./cmd/find.go
./cmd/move.go
./cmd/replace.go
./cmd/root.go
./cmd/search.go
./main.go
```

### move

`gosr move` expands on `gosr find` by adding a second argument for a replacement pattern. Move finds files that match the regex just like `gosr find` but then moves the files (or copies them if `-c` is provided) to a replacement path determined by the replacement pattern. The replacement pattern can use matched groups like `$0` (entire matched string), `$1` (first matching group), etc. Named matches can also be created using this syntax: `(?P<name>...)` which can then be referenced using `${name}`. Note you can also use `${1}` for index-based match groups, and often should as it can resolve ambiguity between the match name versus the rest of the replacement expression (match names without `{}` will match the longest name possible). `$$` can be used to use a literal `$` character in the replacement string.

```bash
$ gosr move '\.go$' '${0}.tmp' -r -c
rename ./cmd/find.go -> ./cmd/find.go.tmp [y/n/Y/N]: N
rename ./cmd/move.go -> ./cmd/move.go.tmp [y/n/Y/N]: N
rename ./cmd/replace.go -> ./cmd/replace.go.tmp [y/n/Y/N]: N
rename ./cmd/root.go -> ./cmd/root.go.tmp [y/n/Y/N]: N
rename ./cmd/search.go -> ./cmd/search.go.tmp [y/n/Y/N]: N
rename ./main.go -> ./main.go.tmp [y/n/Y/N]: N
```

The `-s` (simple) flag can be used to make sure the output is only the new filenames to match the output you would get from `gosr find`, which allows a move to be chained into a search or replace (described below).

### search

`gosr search` searches given files for text which matches an expression. The files it searches can be passed as arguments (after the search expression) or passed in via stdin by named pipe.

```bash
$ gosr search 'make\(.*?\)' ./cmd/find.go
./cmd/find.go:28                                items := make(chan string)
./cmd/find.go:29                                error := make(chan error, 1) // buffered so the error can wait until we're done processing data
```

### replace

`gosr replace` expands on `gosr search` by adding a second argument for a replacement pattern. Any file processed (passed as argument or from named pipe) will be rewritten by replacing matching lines per the replacement pattern. NOTE: all files will be rewritten, even if there are no matches, so make sure only files you intend to rewrite are included.

```bash
$ gosr replace 'confirm\((.*)\)$' 'confirm(new_default_argument, ${1})' ./cmd/replace.go 
./cmd/replace.go:85
                                confirmation, err := confirm("%v:%v replace\n\t%v\n\t->\n\t%v\n?", path, lineNum, line, newLine)
        ->
                                confirmation, err := confirm(new_default_argument, "%v:%v replace\n\t%v\n\t->\n\t%v\n?", path, lineNum, line, newLine)
? [y/n/Y/N]: N
```

## Chaining GOSR

The fact `gosr search` and `gosr replace` can collect filenames from a named pipe means we can chain gosr commands together to match text in certain files, for example in a C++ project you could rename all header files from .h to .hpp and simultanously rewrite the includes with a command like this:
`gosr move '.*inc/(.*?)\.h$' 'inc/$1.hpp' -s -r | gosr replace '#include "inc/(.*?)\.h" '#include "inc/$1.hpp"'`

## Common Flags

`-q` / `--quiet` is quiet mode and will suppress all output (besides user prompts)
`-y` / `--confirm` will suppress confirmation and apply all changes without prompting, combined with `-q` this will silently make all changes with no output at all
`-d` / `--dry` will only show would-be changes without applying any changes. This is essentially the opposite of `-y` with which it is mutually exclusive. `-d` is also mutually exclusive with `-q` as such a command would do nothing.

For `search` and `replace`, which by default expect STDIN to provide a list of files (if piped in), it may instead be required to treat data from STDIN as text to search or replace.
`-i` / `--stdin` accomplishes this, telling GOSR to treat stdin as text for searching and replacing. In the case of `gosr replace` the replacement text will be written to STDOUT and will always be performed regardless of the whether or not `-d` was provided (nor if `N` was used to answer a prompt)

Note the difference between these two chained commands:

```bash
$ gosr find '\.go$' -r | gosr search 'move'
./cmd/move.go:14                moveCommand.Flags().BoolVarP(&recurse, "recursive", "r", false, "Search for matching files recursively in subdirectories")
./cmd/move.go:15                moveCommand.Flags().BoolVarP(&simpleOutput, "simple", "s", false, "Simple output - omit original filenames, implies --confirm. This allows chaining a move into another command like search or replace")
./cmd/move.go:16                moveCommand.Flags().BoolVarP(&byCopy, "copy", "c", false, "Copy files to destination instead of moving")
./cmd/move.go:18                rootCmd.AddCommand(moveCommand)
./cmd/move.go:24                moveCommand  = &cobra.Command{
./cmd/move.go:25                        Use:   "move <FILE PATTERN> <REPLACEMENT PATTERN>",
./cmd/move.go:26                        Short: "move files matching a regex pattern to a new replaced filepath",
./cmd/move.go:76                                                        // now move/copy the actual file
./cmd/root.go:32                // gosr move <regex> <replace_regex>
./cmd/root.go:40                // gosr move '.*\.h$' '$1.hpp' -s | gosr search '#include ".*?\.h" '#include "$1.hpp"'
```

versus

```bash
$ gosr find '\.go$' -r | gosr search 'move' -i
./cmd/move.go
```

## User Prompts

By default (without `-y`) GOSR will confirm changes before applying them, and will request outlike like `[y/n/Y/N]`. y and n obviously stand for yes and no respectively, but the uppercase Y and N will provide that answer for all subsequent prompts. This means once you've answered `Y` GOSR will continue as if `-y` had been provided, and if you instead had answered with `N` GOSR continues as if `-d` were instead provided.

## Regex Syntax

GOSR uses go's regexp package, complete documentation on the syntax can be found [here](https://github.com/google/re2/wiki/Syntax)

## Replacement Syntax

GOSR's replacement syntax ultimately depends on the regexp package's `Expand` functionality, which is documented [here](https://pkg.go.dev/regexp#Regexp.Expand)