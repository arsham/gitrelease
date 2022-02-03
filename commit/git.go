// Package commit contains the logic for interacting with git, commits and
// github.
package commit

import (
	"context"
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
