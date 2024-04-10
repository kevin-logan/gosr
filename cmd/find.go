package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

func addFindCommand(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&recurse, "recursive", "r", false, "Search for matching files recursively in subdirectories")
	cmd.Flags().BoolVarP(&flip, "flip", "f", false, "Flip results: only output non-matches")
	rootCmd.AddCommand(cmd)
}

func init() {
	addFindCommand(findCommand)
	addFindCommand(findCppCommand)
	addFindCommand(findGoCommand)
	addFindCommand(findPhpCommand)
	addFindCommand(findPythonCommand)
	addFindCommand(findRustCommand)
	addFindCommand(findXmlCommand)
	addFindCommand(findJsonCommand)
	addFindCommand(findJavaCommand)
	addFindCommand(findJsCommand)
}

var (
	recurse     = false
	flip        = false
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
	findCppCommand = &cobra.Command{
		Use:   "find-cpp",
		Short: "find files matching a preprogrammed C++ file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.(h|hh|hpp|cpp|cxx|cc|c|mxx|tcc|txx)$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findGoCommand = &cobra.Command{
		Use:   "find-go",
		Short: "find files matching a preprogrammed Go file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.go$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findPhpCommand = &cobra.Command{
		Use:   "find-php",
		Short: "find files matching a preprogrammed PHP file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.(php|inc)$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findPythonCommand = &cobra.Command{
		Use:   "find-python",
		Short: "find files matching a preprogrammed Python file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.py$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findRustCommand = &cobra.Command{
		Use:   "find-rust",
		Short: "find files matching a preprogrammed Rust file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.rs$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findXmlCommand = &cobra.Command{
		Use:   "find-xml",
		Short: "find files matching a preprogrammed XML file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.xml$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findJsonCommand = &cobra.Command{
		Use:   "find-json",
		Short: "find files matching a preprogrammed JSON file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.json$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findJavaCommand = &cobra.Command{
		Use:   "find-java",
		Short: "find files matching a preprogrammed Java file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.java$`)
			return findCommand.RunE(cmd, args)
		},
	}
	findJsCommand = &cobra.Command{
		Use:   "find-js",
		Short: "find files matching a preprogrammed JavaScript file matching regex pattern",
		Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(args, `\.js$`)
			return findCommand.RunE(cmd, args)
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
		matches := re.MatchString(name)
		if matches == !flip {
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
