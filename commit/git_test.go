package commit_test

import (
	"context"
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
