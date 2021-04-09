# gode

`gode` is a small utility script for automating code modifications using regex patterns.

[regexp2](https://github.com/dlclark/regexp2) in RE2 compatability mode is used for the regexp engine, allowing for more complex patterns.

## Installation

```bash
go get -u github.com/meanguy/gode/bin/gode
```

## Usage

```bash
$ gode -h
usage: gode: [flags] [path ...]
  -d    display diffs instead of rewriting files
  -debug
        enable debug mode
  -filetype value
        specify filetype to apply modifications to
  -m string
        regex pattern to match lines for modifying
  -r string
        regex pattern to replace matched lines
  -recursive
        enable recursive file search
  -w    write result to source file instead of stdout
```

## Examples

```bash
$ cat example.go
package main

import (
        "errors"
        "flag"
        "fmt"
        "os"
)

var ErrBadArgument = errors.New("bad program argument")

func program(argument string) error {
        if argument != "g'day" {
                return ErrBadArgument
        }

        return nil
}

func main() {
        flag.Parse()

        if flag.NArg() < 1 {
                fmt.Fprintf(os.Stderr, "usage: %s [argument]\n", os.Args[0])
                os.Exit(-1)
        }

        err := program(os.Args[1])
        if err != nil {
                if err == ErrBadArgument {
                        fmt.Fprint(os.Stderr, "try another argument\n")
                }

                fmt.Fprintf(os.Stderr, "unknown error: %v", err)
        }
}

# Update Go error checking to use Go 1.13 library features.
$ gode -m 'err (!?)={1,2} ((?!nil)[A-Za-z.]+)' -r '${1}errors.Is(err, ${2})' example.go
package main

import (
        "errors"
        "flag"
        "fmt"
        "os"
)

var ErrBadArgument = errors.New("bad program argument")

func program(argument string) error {
        if argument != "g'day" {
                return ErrBadArgument
        }

        return nil
}

func main() {
        flag.Parse()

        if flag.NArg() < 1 {
                fmt.Fprintf(os.Stderr, "usage: %s [argument]\n", os.Args[0])
                os.Exit(-1)
        }

        err := program(os.Args[1])
        if err != nil {
                if errors.Is(err, ErrBadArgument) {
                        fmt.Fprint(os.Stderr, "try another argument\n")
                }

                fmt.Fprintf(os.Stderr, "unknown error: %v", err)
        }
}

# Print a colored diff view instead.
$ gode -m 'err (!?)={1,2} ((?!nil)[A-Za-z.]+)' -r '${1}errors.Is(err, ${2})' -d example.go
package main

import (
        "errors"
        "flag"
        "fmt"
        "os"
)

var ErrBadArgument = errors.New("bad program argument")

func program(argument string) error {
        if argument != "g'day" {
                return ErrBadArgument
        }

        return nil
}

func main() {
        flag.Parse()

        if flag.NArg() < 1 {
                fmt.Fprintf(os.Stderr, "usage: %s [argument]\n", os.Args[0])
                os.Exit(-1)
        }

        err := program(os.Args[1])
        if err != nil {
                if err ==ors.Is(err, ErrBadArgument) {
                        fmt.Fprint(os.Stderr, "try another argument\n")
                }

                fmt.Fprintf(os.Stderr, "unknown error: %v", err)
        }
}

# Write changes to the source file.
$ gode -m 'err (!?)={1,2} ((?!nil)[A-Za-z.]+)' -r '${1}errors.Is(err, ${2})' -w example.go

$ cat example.go 
package main

import (
        "errors"
        "flag"
        "fmt"
        "os"
)

var ErrBadArgument = errors.New("bad program argument")

func program(argument string) error {
        if argument != "g'day" {
                return ErrBadArgument
        }

        return nil
}

func main() {
        flag.Parse()

        if flag.NArg() < 1 {
                fmt.Fprintf(os.Stderr, "usage: %s [argument]\n", os.Args[0])
                os.Exit(-1)
        }

        err := program(os.Args[1])
        if err != nil {
                if errors.Is(err, ErrBadArgument) {
                        fmt.Fprint(os.Stderr, "try another argument\n")
                }

                fmt.Fprintf(os.Stderr, "unknown error: %v", err)
        }
}

# Recursively apply modifications to any directory arguments.
$ gode -recursive -m 'err (!?)={1,2} ((?!nil)[A-Za-z.]+)' -r '${1}errors.Is(err, ${2})' -w .

# Filter out any files not matching a filetype -- useful with -recursive.
$ gode -filetype '.txt' -recursive -m 'err (!?)={1,2} ((?!nil)[A-Za-z.]+)' -r '${1}errors.Is(err, ${2})' -d .

# Only modify files matching a set of filetypes.
$ gode -filetype '.py' -filetype '.go' -recursive -m 'TODO:' -r 'TODO(XX):' -w ~/docs/dev/py ~/docs/dev/golang
```
