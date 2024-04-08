package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

func init() {
	findCommand.Flags().BoolVarP(&recurse, "recursive", "r", false, "Search for matching files recursively in subdirectories")
	rootCmd.AddCommand(findCommand)
}

var (
	recurse     = false
	findCommand = &cobra.Command{
		Use:   "find <FILE PATTERN>",
		Short: "find files matching a regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			re, err := regexp.Compile(args[0])
			if err != nil {
				return err
			}

			items, error := gosrWalk(".", re, recurse)

		pollLoop:
			for {
				select {
				case err = <-error:
					// done as soon as we hit an error
					return err
				case item, ok := <-items:
					if !ok {
						// channel closed, we're done
						break pollLoop
					}

					output("%v\n", item)
				}
			}

			// success if we reached here without error
			return nil
		},
	}
)

func gosrWalk(path string, re *regexp.Regexp, recurse bool) (chan string, chan error) {
	out := make(chan string)
	errorOut := make(chan error)

	go func() {
		// ensure output is closed when we're done so reader can know when to stop
		defer close(out)
		defer close(errorOut)

		err := walkImpl(path, re, recurse, out)

		if err != nil {
			errorOut <- err
		}
	}()

	return out, errorOut
}

func walkImpl(path string, re *regexp.Regexp, recurse bool, out chan string) error {
	items, err := os.ReadDir(path)

	if err != nil {
		return err
	}

	for _, item := range items {
		name := fmt.Sprintf("%v/%v", path, item.Name())
		if re.MatchString(name) {
			out <- name
		}

		if item.IsDir() && recurse {
			err = walkImpl(name, re, recurse, out)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
