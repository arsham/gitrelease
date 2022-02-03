// Package commit contains the logic for interacting with git, commits and
// github.
package commit

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Git executes git processes targeted at a directory. If the Dir property is
// empty, all calls will be on the current folder.
type Git struct {
	Dir string
}

// LatestTag returns the last tag in the repository.
func (g Git) LatestTag(ctx context.Context) (string, error) {
	args := []string{
		"describe",
		"--tags",
		"--abbrev=0",
	}
	// nolint:gosec // we don't have any other way to get the previous tag.
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, string(out))
	}

	return strings.Trim(string(out), "\n"), nil
}

// PreviousTag returns the previous tag of the given tag.
func (g Git) PreviousTag(ctx context.Context, tag string) (string, error) {
	args := []string{
		"describe",
		"--tags",
		"--abbrev=0",
		tag + "^",
	}
	// nolint:gosec // we don't have any other way to get the previous tag.
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, string(out))
	}

	return strings.Trim(string(out), "\n"), nil
}

// Commits returns the contents of all commits between two tags.
func (g Git) Commits(ctx context.Context, tag1, tag2 string) ([]string, error) {
	separator := "00000000000000000000000000000000000"
	args := []string{
		"log",
		"--oneline",
		fmt.Sprintf("%s..%s", tag1, tag2),
		fmt.Sprintf("--pretty=%s%%B", separator),
	}
	// nolint:gosec // we need these variables.
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, string(out))
	}
	logs := strings.Split(string(out), separator)
	return logs, nil
}
