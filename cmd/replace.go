package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

func init() {
	replaceCommand.Flags().BoolVarP(&stdinAsText, "stdin", "i", false, "Treat stdin as text to replace (versus a list of files). Replacement written to STDOUT.")

	rootCmd.AddCommand(replaceCommand)
}

var replaceCommand = &cobra.Command{
	Use:   "replace <PATTERN> <REPLACEMENT PATTERN> [FILE]...",
	Short: "search files provided (or files from STDIN) for lines matching a regex pattern and replace with the given replacement pattern",
	Args:  cobra.MatchAll(cobra.MinimumNArgs(2), cobra.OnlyValidArgs),
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

		if !stdinAsText && info.Mode()&os.ModeNamedPipe != 0 {
			// read lines into filelist
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				filelist = append(filelist, scanner.Text())
			}
		}

		// read args into filelist, start at 2 to skip pattern and replacement pattern
		for i := 2; i < len(args); i++ {
			filelist = append(filelist, args[i])
		}

		for _, path := range filelist {
			err = replaceLines(path, re, args[1])
			if err != nil {
				return err
			}
		}

		// finally handle stdinAsText
		if stdinAsText {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if re.MatchString(line) {
					// no confirmation or checks on STDIN -> STDOUT
					newLine := re.ReplaceAllString(line, args[1])
					fmt.Println(newLine)
				} else {
					fmt.Println(line)
				}
			}
		}

		return nil
	},
}

func replaceLines(path string, re *regexp.Regexp, replacePattern string) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	tmpPath := path + ".gosr.tmp"
	_, err = os.Stat(tmpPath)
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("gosr temporary file [%v] already exists, review and consider deleting the file", tmpPath)
	}

	outFile, err := os.Create(tmpPath)

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			newLine := re.ReplaceAllString(line, replacePattern)
			confirmation, err := confirm("%v:%v\n\t%v\n\t->\n\t%v\n?", path, lineNum, line, newLine)
			if err != nil {
				return err
			}

			if confirmation {
				outFile.WriteString(newLine)
			} else {
				outFile.WriteString(line)
			}
		} else {
			// append raw line
			outFile.WriteString(line)
		}

		// the line doesn't include the newline, append it here
		// should this support non-LF newlines? Go doesn't seem to by default with Println-like functions
		outFile.WriteString("\n")
	}

	// now replace original file iwth tmp
	return os.Rename(tmpPath, path)
}
