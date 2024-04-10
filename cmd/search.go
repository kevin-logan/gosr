package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

func init() {
	searchCommand.Flags().BoolVarP(&stdinAsText, "stdin", "i", false, "Treat stdin as text to search (versus a list of files to search)")
	searchCommand.Flags().BoolVarP(&flip, "flip", "f", false, "Flip results: only output non-matches")

	rootCmd.AddCommand(searchCommand)
}

type searchData struct {
	items chan string
	error chan error
}

var (
	stdinAsText   = false
	searchCommand = &cobra.Command{
		Use:   "search <PATTERN> [FILE]...",
		Short: "search files provided (or files from STDIN) for lines matching a regex pattern",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get pattern regex first
			re, err := regexp.Compile(args[0])
			if err != nil {
				return err
			}

			// files come from `args`, and/or from STDIN
			filelist := make([]string, 0, len(args)-1)
			info, err := os.Stdin.Stat()
			if err != nil {
				return err
			}

			// if stdin isn't search text, check it for additional files to search
			if !stdinAsText && info.Mode()&os.ModeNamedPipe != 0 {
				// read lines into filelist
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					filelist = append(filelist, scanner.Text())
				}
			}

			// read args into filelist, start at 1 to skip pattern
			for i := 1; i < len(args); i++ {
				filelist = append(filelist, args[i])
			}

			// channels with enough buffer for each
			error := make(chan error, len(filelist))
			items := make([]*searchData, 0, len(filelist))

			for _, path := range filelist {
				item := searchData{make(chan string), error}

				go readLines(path, re, &item)

				items = append(items, &item)
			}

			if stdinAsText {
				item := searchData{make(chan string), error}
				go readLinesFromFile(os.Stdin, nil, re, &item)
				items = append(items, &item)
			}

			for _, v := range items {
			pollLoop:
				for {
					select {
					case err = <-error:
						if err != nil {
							// just bail as soon as we have an error
							return err
						}
					case lineInfo, ok := <-v.items:
						if !ok {
							// channel is closed, we've read all input
							break pollLoop
						}

						output("%v\n", lineInfo)
					}
				}
			}

			// we've processed all data, do one last check for an error
			select {
			case err = <-error:
				return err
			default:
				return nil
			}
		},
	}
)

func readLines(path string, re *regexp.Regexp, item *searchData) {
	file, err := os.Open(path)

	if err != nil {
		item.error <- err
		close(item.items)
		return
	}

	readLinesFromFile(file, &path, re, item)
}

func readLinesFromFile(file *os.File, path *string, re *regexp.Regexp, item *searchData) {
	defer close(item.items)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := re.MatchString(line)

		if matches == !flip {
			if path == nil {
				item.items <- line
			} else {
				item.items <- fmt.Sprintf("%v:%v\t%v", *path, lineNum, line)
			}
		}
	}
}
