package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "gosr",
		Short: "A flexible find, rename, search, and replace tool in Go",
		Long: `gosr is a command line interface that supports searching for files
using regex, searching within those (or all) files via regex, while potentially
renaming the files and replacing the contents with support for regex capture
groups.`,
	}
	quiet      bool = false
	confirmAll bool = false
	dryRun     bool = false
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "supress output")
	rootCmd.PersistentFlags().BoolVarP(&confirmAll, "confirm", "y", false, "confirm all prompts as if responding Y")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry", "d", false, "dry run: don't perform any changes, as if responding to prompts with N")
	rootCmd.MarkFlagsMutuallyExclusive("dry", "confirm")
	rootCmd.MarkFlagsMutuallyExclusive("dry", "quiet")
}

func output(format string, a ...any) {
	if !quiet {
		fmt.Printf(format, a...)
	}
}

func confirm(format string, a ...any) (bool, error) {
	format += " [y/n/Y/N]: "
	if dryRun {
		// only output the question for the dry run
		output(format+"N\n", a...)
		return false, nil
	} else {
		if confirmAll {
			// use output if confirmAll, so if we're quiet we don't print anything on -q -y
			output(format+"Y\n", a...)
			return true, nil
		}

		// we need confirmation, so we _can't_ use `output` as we don't want output suppressed with -q so user sees prompt
		fmt.Printf(format, a...)
		var choice string

		// open /dev/tty directly as we expect to often have our stdin piped from chained commands
		in, err := os.Open("/dev/tty")
		if err != nil {
			return false, err
		}

		_, err = fmt.Fscanln(in, &choice)
		in.Close()
		if err != nil {
			return false, err
		}

		switch choice {
		case "y":
			return true, nil
		case "n":
			return false, nil
		case "Y":
			confirmAll = true
			return true, nil
		case "N":
			dryRun = true
			return false, nil
		default:
			return false, fmt.Errorf("invalid response [%v], only y/n/Y/N answers are accepted", choice)
		}
	}
}
