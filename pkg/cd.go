package pkg

import (
	"errors"
	"fmt"
	"os"

	"github.com/grdl/git-get/pkg/git"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sahilm/fuzzy"
)

var (
	ErrNoMatchFound         = errors.New("no repository matching found")
	ErrMissingQueryTerminal = errors.New("missing query for non-interactive terminal")
)

// CdCfg provides configuration for the Cd command.
type CdCfg struct {
	Roots []string
	Query string
}

// Cd executes the "git cd" command.
func Cd(conf *CdCfg) error {
	finder := git.NewRepoFinder(conf.Roots)
	if err := finder.Find(); err != nil {
		return err
	}

	repos := finder.LoadAll(false)

	paths := make([]string, len(repos))
	for i, r := range repos {
		paths[i] = r.Path()
	}

	var selectedPath string

	var err error

	if conf.Query != "" {
		selectedPath, err = findSelectedPathWithQuery(conf.Query, paths)
	} else {
		selectedPath, err = findSelectedPathNoQuery(paths)
	}

	if err != nil {
		return err
	}

	fmt.Println(selectedPath)

	return nil
}

func findSelectedPathWithQuery(query string, paths []string) (string, error) {
	matches := fuzzy.Find(query, paths)
	if len(matches) == 0 {
		return "", fmt.Errorf("%w: '%s'", ErrNoMatchFound, query)
	}

	if len(matches) == 1 {
		return paths[matches[0].Index], nil
	}

	// Multiple matches, if interactive use fuzzyfinder
	if isInteractive() {
		return findInteractive(matches, paths)
	}

	// Non-interactive: pick the best match (the first one from fuzzy.Find).
	return paths[matches[0].Index], nil
}

func findSelectedPathNoQuery(paths []string) (string, error) {
	if !isInteractive() {
		return "", ErrMissingQueryTerminal
	}

	idx, err := fuzzyfinder.Find(
		paths,
		func(i int) string {
			return paths[i]
		},
	)
	if err != nil {
		return "", fmt.Errorf("fuzzyfinder failed: %w", err)
	}

	return paths[idx], nil
}

func findInteractive(matches fuzzy.Matches, paths []string) (string, error) {
	// Filter paths to matches only for fuzzyfinder to start with
	filteredPaths := make([]string, len(matches))
	for i, m := range matches {
		filteredPaths[i] = paths[m.Index]
	}

	idx, err := fuzzyfinder.Find(
		filteredPaths,
		func(i int) string {
			return filteredPaths[i]
		},
	)
	if err != nil {
		return "", fmt.Errorf("fuzzyfinder failed: %w", err)
	}

	return filteredPaths[idx], nil
}

func isInteractive() bool {
	fileInfo, _ := os.Stdin.Stat()

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
