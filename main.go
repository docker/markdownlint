package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SvenDowideit/markdownlint/checkers"
	"github.com/SvenDowideit/markdownlint/data"
	"github.com/SvenDowideit/markdownlint/linereader"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(-1)
	}
	dir := args[0]
	filter := ""
	if len(args) >= 2 {
		filter = args[1]
	}

	data.AllFiles = make(map[string]*data.FileDetails)

	fmt.Println("Finding files")
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return err
		}
		data.VerboseLog("FOUND: %s\n", path)
		if info.IsDir() {
			return nil
		}
		file, err := filepath.Rel(dir, path)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return err
		}
		// verboseLog("\t walked to %s\n", file)
		data.AddFile(file, path)
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(-1)
	}

	for file, details := range data.AllFiles {
		if !strings.HasPrefix(file, filter) {
			data.VerboseLog("FILTERED: %s\n", file)
			continue
		}
		if !strings.HasSuffix(file, ".md") {
			data.VerboseLog("SKIPPING: %s\n", file)
			continue
		}
		fmt.Printf("opening: %s\n", file)

		reader, err := linereader.OpenReader(details.FullPath)
		if err != nil {
			fmt.Printf("ERROR opening: %s\n", err)
			data.AllFiles[file].FormatErrorCount++
		}

		err = checkers.CheckHugoFrontmatter(reader, file)
		if err != nil {
			fmt.Printf("ERROR (%s) frontmatter: %s\n", file, err)
		}

	    if draft, ok := data.AllFiles[file].Meta["draft"]; ok || draft == "true" {
            fmt.Printf("Draft=%s: SKIPPING %s link check.\n", draft, file)
        } else {
            //fmt.Printf("Draft=%s: %s link check.\n", draft, file)
            err = checkers.CheckMarkdownLinks(reader, file)
            if err != nil {
                // this only errors if there is a fatal issue
                fmt.Printf("ERROR (%s) links: %s\n", file, err)
                data.AllFiles[file].FormatErrorCount++
            }
        }
		reader.Close()
	}
	checkers.TestLinks()

	// TODO (JIRA: DOCS-181): Title, unique across products if not, file should include an {identifier}

	summaryFileName := "markdownlint.summary.txt"
	f, err := os.Create(summaryFileName)
	if err == nil {
		fmt.Printf("Also writing summary to %s :\n\n", summaryFileName)
		defer f.Close()
	}

	if filter != "" {
		Printf(f, "# Filtered (%s) Summary:\n\n", filter)
	} else {
		Printf(f, "# Summary:\n\n")
	}
	errorCount, errorString := checkers.FrontSummary(filter)
	Printf(f, errorString)
	count, errorString := checkers.LinkSummary(filter)
	errorCount += count
	Printf(f, errorString)
	Printf(f, "\n\tFound: %d files\n", len(data.AllFiles))
	Printf(f, "\tFound: %d errors\n", errorCount)
	// return the number of 404's to show that there are things to be fixed
	os.Exit(errorCount)
}

func Printf(f *os.File, format string, a ...interface{}) {
	str := fmt.Sprintf(format, a...)
	fmt.Print(str)
	if f != nil {
		// Don't reall want to know we can't write..
		f.WriteString(str)
	}
}

func printUsage() {
	fmt.Println("Please specify a directory to check")
	fmt.Println("\tfor example: markdownlint . [filter]")
	fmt.Println("\t [filter] can be any string prefix inside the dir specified")
	flag.PrintDefaults()
}
