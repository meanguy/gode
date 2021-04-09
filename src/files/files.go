package files

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/apex/log"
)

var (
	ErrRecursionLimitReached = errors.New("reached maximum recursion depth")

	RecurseLimit = 128
)

type FilterFn func(string) bool

func FindRecursive(ctx context.Context, directories []string, filters ...FilterFn) ([]string, error) {
	return findRecursiveImpl(ctx, 0, directories, filters...)
}

func findRecursiveImpl(ctx context.Context, depth int, directories []string, filters ...FilterFn) ([]string, error) {
	logger := log.FromContext(ctx)

	if len(directories) == 0 {
		return []string{}, nil
	}

	if depth >= RecurseLimit {
		logger.WithField("depth", depth).Debug("hit recusion limit")
		return nil, ErrRecursionLimitReached
	}

	matches := []string{}
	subdirs := []string{}
	for _, directory := range directories {
		subpaths, err := os.ReadDir(directory)
		if err != nil {
			return nil, err
		}

		for _, subpath := range subpaths {
			path := fmt.Sprintf("%s/%s", directory, subpath.Name())

			if subpath.IsDir() {
				logger.WithFields(log.Fields{
					"depth":  depth,
					"subdir": path,
				}).Debug("found new directory")
				subdirs = append(subdirs, path)
			} else {
				logger.WithFields(log.Fields{
					"depth":    depth,
					"filepath": path,
				}).Debug("found new file")
				matches = append(matches, path)
			}
		}
	}

	recursed, err := findRecursiveImpl(ctx, depth+1, subdirs, filters...)
	if err != nil {
		return nil, err
	}

	for _, file := range recursed {
		for _, filter := range filters {
			if filter(file) {
				matches = append(matches, file)
				break
			}
		}
	}

	return matches, nil

}
