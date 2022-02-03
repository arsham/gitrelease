package commit_test

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func appendToFile(t *testing.T, dir, filename, msg string) {
	t.Helper()
	require.NoError(t, os.Chdir(dir))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	require.NoError(t, err)
	defer f.Close()

	_, err = f.WriteString(msg)
	require.NoError(t, err)
}

var cmpIgnoreNewlines = cmp.Transformer("IgnoreNewlines", func(in string) string {
	return strings.ReplaceAll(in, "\n", "")
})

var stringSliceCleaner = cmp.Transformer("CleanStringSlice", func(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
})

var commitComparer = cmp.Options{
	cmpIgnoreNewlines,
	stringSliceCleaner,
}
