package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/dlclark/regexp2"
	"github.com/meanguy/gode/src/files"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type args struct {
	Debug     bool
	Recursive bool
	DiffMode  bool
	WriteMode bool

	MatchPattern   string
	ReplacePattern string

	Filetypes   []string
	SearchPaths []string
}

func parseArgs() args {
	opts := args{}
	flag.BoolVar(&opts.Debug, "debug", false, "enable debug mode")
	flag.BoolVar(&opts.Recursive, "recursive", false, "enable recursive file search")
	flag.BoolVar(&opts.DiffMode, "d", false, "display diffs instead of rewriting files")
	flag.BoolVar(&opts.WriteMode, "w", false, "write result to source file instead of stdout")

	flag.StringVar(&opts.MatchPattern, "m", "", "regex pattern to match lines for modifying")
	flag.StringVar(&opts.ReplacePattern, "r", "", "regex pattern to replace matched lines")

	flag.Func("filetype", "specify filetype to apply modifications to", func(s string) error {
		opts.Filetypes = append(opts.Filetypes, s)

		return nil
	})

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s: [flags] [path ...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(-1)
	}

	opts.SearchPaths = flag.Args()
	if len(opts.MatchPattern) == 0 || len(opts.ReplacePattern) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	return opts
}

//nolint:cyclop,funlen,gocognit // TODO: Need to refactor
func main() {
	logger := log.Log
	ctx := log.NewContext(context.Background(), logger)
	opts := parseArgs()

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	filters := []files.FilterFn{}
	for _, filetype := range opts.Filetypes {
		filters = append(filters, func(file string) bool {
			return strings.HasSuffix(file, filetype)
		})
	}

	if len(filters) == 0 {
		// Match all files if there's no other filters.
		filters = append(filters, func(s string) bool {
			return true
		})
	}

	searchDirectories := []string{}
	searchFiles := []string{}

	for _, path := range opts.SearchPaths {
		stat, err := os.Stat(path)

		switch {
		case err != nil:
			logger.Fatal(err.Error())
		case stat.IsDir():
			if !opts.Recursive {
				logger.Fatalf("cannot search directory without -recursive")
			}

			searchDirectories = append(searchDirectories, path)
		default:
			searchFiles = append(searchFiles, path)
		}
	}

	matchPattern, err := regexp2.Compile(opts.MatchPattern, regexp2.RE2)
	if err != nil {
		logger.Fatal(err.Error())
	}

	matches := append([]string(nil), searchFiles...)

	if opts.Recursive {
		recursed, err := files.FindRecursive(ctx, searchDirectories)
		if err != nil {
			logger.Fatal(err.Error())
		}

		matches = append(matches, recursed...)
	}

	for _, match := range matches {
		skip := true

		for _, filter := range filters {
			if filter(match) {
				skip = false

				break
			}
		}

		if skip {
			continue
		}

		file, err := os.Stat(match)
		if err != nil {
			logger.Fatal(err.Error())
		}

		writer, err := os.OpenFile(match, os.O_RDWR, file.Mode().Perm())
		if err != nil {
			logger.Fatal(err.Error())
		}
		defer writer.Close()

		raw, err := ioutil.ReadAll(writer)
		if err != nil {
			logger.Fatal(err.Error())
		}

		original := string(raw)
		if matched, err := matchPattern.MatchString(original); err != nil {
			logger.Fatal(err.Error())
		} else if !matched {
			logger.WithField("filepath", match).Debug("no pattern match")

			continue
		}

		replaced, err := matchPattern.Replace(original, opts.ReplacePattern, -1, -1)
		if err != nil {
			logger.Fatal(err.Error())
		}

		switch {
		case opts.WriteMode:
			truncated, err := os.OpenFile(match, os.O_WRONLY|os.O_TRUNC, file.Mode().Perm())
			if err != nil {
				logger.Fatal(err.Error())
			}
			defer truncated.Close()

			if _, err := fmt.Fprint(truncated, replaced); err != nil {
				logger.WithField("filepath", match).Fatal(err.Error())
			}
		case opts.DiffMode:
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffCleanupSemantic(
				dmp.DiffCleanupEfficiency(dmp.DiffMain(original, replaced, true)))

			if _, err := fmt.Fprint(os.Stdout, dmp.DiffPrettyText(diffs)); err != nil {
				logger.Fatal(err.Error())
			}
		default:
			if _, err := fmt.Fprint(os.Stdout, replaced); err != nil {
				logger.Fatal(err.Error())
			}
		}
	}
}
