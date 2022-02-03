package commit_test

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func createGitRepo(t *testing.T) string {
	t.Helper()
	dir, err := ioutil.TempDir("", "git-lfs-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})

	newDir := path.Join(dir, "project")
	os.Mkdir(newDir, 0o755)

	commands := [][]string{
		{"init"},
		{"config", "user.email", "arsham@github.com"},
		{"config", "user.name", "arsham"},
	}

	for _, args := range commands {
		cmd := exec.CommandContext(context.Background(), "git", args...)
		cmd.Dir = newDir
		_, err = cmd.CombinedOutput()
		require.NoError(t, err)
	}

	return newDir
}

func createGitTag(t *testing.T, dir, tag string) {
	t.Helper()
	args := []string{"tag", tag}
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func createFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	require.NoError(t, os.Chdir(dir))
	require.NoError(t, os.WriteFile(filename, []byte(content), 0o644))
}

func commitChanges(t *testing.T, dir, msg string) {
	t.Helper()

	args := []string{"add", "-A"}
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	args = []string{"commit", "-am", msg, "--no-gpg-sign"}
	cmd = exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}
