// Package commit contains the logic for interacting with git, commits and
// github.
package commit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/github-release/github-release/github"
	"github.com/pkg/errors"
)

const baseURL = "https://api.github.com"

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

var infoRe = regexp.MustCompile(`github\.com[:/](?P<user>[^/]+)/(?P<repo>[^\n.]+)(\.git)?`)

// RepoInfo returns some information about the repository.
func (g Git) RepoInfo(ctx context.Context) (user, repo string, err error) {
	args := []string{
		"config",
		"--get",
		"remote.origin.url",
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Wrap(err, string(out))
	}

	info := infoRe.FindStringSubmatch(string(out))
	if len(info) != 4 {
		return "", "", fmt.Errorf("could not parse repository info: %s", string(out))
	}
	user = info[1]
	repo = info[2]

	return user, repo, nil
}

type releaseCreate struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish,omitempty"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

// Release publishes the release for the user on the repo.
func (g Git) Release(token, user, repo, tag, desc string) error {
	params := releaseCreate{
		TagName: tag,
		Body:    desc,
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "marshalling values")
	}

	client := github.NewClient(repo, token, nil)
	client.SetBaseURL(baseURL)
	reader := bytes.NewReader(payload)
	req, err := client.NewRequest("POST", fmt.Sprintf("/repos/%s/%s/releases", user, repo), reader)
	if err != nil {
		return errors.Wrapf(err, "creating request to the API: %q", string(payload))
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "submitting to the API: %q", string(payload))
	}
	// nolint:errcheck // it's ok.
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == 422 {
			return errors.New("release already exists")
		}
		return fmt.Errorf("error publishing release with code: %q", resp.Status)
	}
	return nil
}
