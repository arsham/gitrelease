package commit_test

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/arsham/gitrelease/commit"
	"github.com/blokur/testament"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	t.Parallel()
	t.Run("LatestTag", testGitLatestTag)
	t.Run("PreviousTag", testGitPreviousTag)
	t.Run("Commits", testGitCommits)
	t.Run("RepoInfo", testGitRepoInfo)
}

func testGitLatestTag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := createGitRepo(t)

	g := commit.Git{
		Dir: dir,
	}

	_, err := g.LatestTag(ctx)
	assert.Error(t, err)

	createFile(t, dir, "file.txt", testament.RandomString(20))
	commitChanges(t, dir, testament.RandomString(20))
	createGitTag(t, dir, "v0.0.1")

	got, err := g.LatestTag(ctx)
	require.NoError(t, err)
	assert.Equal(t, "v0.0.1", got)

	createFile(t, dir, "file2.txt", testament.RandomString(20))
	commitChanges(t, dir, testament.RandomString(20))
	createGitTag(t, dir, "v0.0.2")

	createFile(t, dir, "file3.txt", testament.RandomString(20))
	commitChanges(t, dir, testament.RandomString(20))
	got, err = g.LatestTag(ctx)
	require.NoError(t, err)
	assert.Equal(t, "v0.0.2", got)
}

func testGitPreviousTag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := createGitRepo(t)

	g := commit.Git{
		Dir: dir,
	}

	_, err := g.PreviousTag(ctx, "v0.0.10")
	assert.Error(t, err)

	createFile(t, dir, "file.txt", testament.RandomString(20))
	commitChanges(t, dir, testament.RandomString(20))
	createGitTag(t, dir, "v0.0.1")

	createFile(t, dir, "file2.txt", testament.RandomString(20))
	commitChanges(t, dir, testament.RandomString(20))
	createGitTag(t, dir, "v0.0.2")

	got, err := g.PreviousTag(ctx, "v0.0.2")
	require.NoError(t, err)
	assert.Equal(t, "v0.0.1", got)

	createFile(t, dir, "file3.txt", testament.RandomString(20))
	commitChanges(t, dir, testament.RandomString(20))
	got, err = g.PreviousTag(ctx, "@")
	require.NoError(t, err)
	assert.Equal(t, "v0.0.2", got)
}

func testGitCommits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := createGitRepo(t)

	g := commit.Git{
		Dir: dir,
	}

	filename := "file.txt"

	createFile(t, dir, filename, testament.RandomString(20))
	commitChanges(t, dir, "msg1")
	createGitTag(t, dir, "v0.0.1")

	msgs := []string{"msg1", "msg2", "msg3"}
	for _, msg := range msgs {
		appendToFile(t, dir, filename, testament.RandomString(20))
		commitChanges(t, dir, msg)
	}

	createGitTag(t, dir, "v0.0.2")

	got, err := g.Commits(ctx, "v0.0.1", "v0.0.2")
	require.NoError(t, err)
	if diff := cmp.Diff(msgs, got, commitComparer...); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func testGitRepoInfo(t *testing.T) {
	t.Parallel()

	wantUser := "arsham666"
	wantRepo := "gitrelease777"
	addrs := map[string]struct {
		addr     string
		wantUser string
		wantRepo string
	}{
		"git protocol":          {"git@github.com:%s/%s", wantUser, wantRepo},
		"git protocol dot":      {"git@github.com:%s/%s", wantUser, wantRepo + ".nvim"},
		"git protocol tail":     {"git@github.com:%s/%s.git", wantUser, wantRepo},
		"git protocol tail dot": {"git@github.com:%s/%s.git", wantUser, wantRepo + ".nvim"},
		"no protocol":           {"github.com/%s/%s", wantUser, wantRepo},
		"no protocol dot":       {"github.com/%s/%s", wantUser, wantRepo + ".nvim"},
		"no protocol tail":      {"github.com/%s/%s.git", wantUser, wantRepo},
		"no protocol tail dot":  {"github.com/%s/%s.git", wantUser, wantRepo + ".nvim"},
		"protocol":              {"https://github.com/%s/%s", wantUser, wantRepo},
		"protocol dot":          {"https://github.com/%s/%s", wantUser, wantRepo + ".nvim"},
		"protocol tail":         {"https://github.com/%s/%s.git", wantUser, wantRepo},
		"protocol tail dot":     {"https://github.com/%s/%s.git", wantUser, wantRepo + ".nvim"},
	}

	for name, tc := range addrs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			dir := createGitRepo(t)
			addr := fmt.Sprintf(tc.addr, tc.wantUser, tc.wantRepo)
			g := commit.Git{
				Dir: dir,
			}
			args := []string{
				"remote",
				"add",
				"origin",
				addr,
			}
			cmd := exec.CommandContext(context.Background(), "git", args...)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, string(out))

			user, repo, err := g.RepoInfo(context.Background())
			require.NoError(t, err, addr)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantRepo, repo)
		})
	}
}
