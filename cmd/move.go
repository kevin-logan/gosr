package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

func init() {
	moveCommand.Flags().BoolVarP(&recurse, "recursive", "r", false, "Search for matching files recursively in subdirectories")
	moveCommand.Flags().BoolVarP(&simpleOutput, "simple", "s", false, "Simple output - omit original filenames, implies --confirm. This allows chaining a move into another command like search or replace")
	moveCommand.Flags().BoolVarP(&byCopy, "copy", "c", false, "Copy files to destination instead of moving")

	rootCmd.AddCommand(moveCommand)
}

var (
	simpleOutput = false
	byCopy       = false
	moveCommand  = &cobra.Command{
		Use:   "move <FILE PATTERN> <REPLACEMENT PATTERN>",
		Short: "move files matching a regex pattern to a new replaced filepath",
		Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			if simpleOutput && dryRun {
				return errors.New("--simple and --dry are mutually exclusive, as --simply implies --confirm")
			}

			re, err := regexp.Compile(args[0])
			if err != nil {
				return err
			}

			items := make(chan string)
			errorOut := make(chan error)
			go gosrWalk(".", re, recurse, items, errorOut)

		pollLoop:
			for {
				select {
				case err = <-errorOut:
					// bail out after first error
					return err
				case item, ok := <-items:
					if !ok {
						break pollLoop
					}
					// for each item attempt replacement to get new path
					newName := re.ReplaceAllString(item, args[1])

					confirmed, err := func() (bool, error) {
						if simpleOutput {
							// simpleOutput implies --confirm
							output("%v\n", newName)
							return true, nil
						} else {
							return confirm("rename %v -> %v", item, newName)
						}
					}()
					if err != nil {
						return err
					}

					if confirmed {
						// first thing is make sure the directory structure up to the new path exists
						newParentDirectory := filepath.Dir(newName)
						err = os.MkdirAll(newParentDirectory, 0755) // this won't do anything if the path already exists
						if err != nil {
							return err
						}

						// now move/copy the actual file
						err = func() error {
							if byCopy {
								return copyFile(item, newName)
							} else {
								return os.Rename(item, newName)
							}
						}()
						if err != nil {
							return err
						}
					}
				}
			}

			// success if we reach here without returning an error previously
			return nil
		},
	}
)

// implementation based on https://stackoverflow.com/a/21067803
func copyFile(src string, dest string) (err error) {
	// open source file
	in, err := os.Open(src)
	if err != nil {
		return
	}

	// we ignore in.Close() error because we don't care, we're not modifying the file
	defer in.Close()

	// create/truncate destination file
	out, err := os.Create(dest)
	if err != nil {
		return
	}

	// defer closing output and potentially overwriting error with Close() error
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	// perform actual data copy
	_, err = io.Copy(out, in)

	if err != nil {
		return
	}

	err = out.Sync()
	return
}
