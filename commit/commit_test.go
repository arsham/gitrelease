package commit_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/arsham/gitrelease/commit"
	"github.com/blokur/testament"
	"github.com/google/go-cmp/cmp"
)

func TestGroupFromCommit(t *testing.T) {
	t.Parallel()
	additional := `\n\n` + testament.RandomString(50)
	tcs := map[string]struct {
		line string
		want commit.Group
	}{
		"not special": {
			line: "something",
			want: commit.NewGroup("Misc", "", "something", false),
		},
		"not special titled": {
			line: "Something",
			want: commit.NewGroup("Misc", "", "Something", false),
		},
		"simply topic": {
			line: "fix something",
			want: commit.NewGroup("Fix", "", "something", false),
		},
		"simply topic multi": {
			line: "fix something" + additional,
			want: commit.NewGroup("Fix", "", "something"+additional, false),
		},
		"simply topic titled": {
			line: "Fix Something",
			want: commit.NewGroup("Fix", "", "Something", false),
		},
		"topic section": {
			line: "Fix(repo) something",
			want: commit.NewGroup("Fix", "repo", "something", false),
		},
		"topic section multi": {
			line: "Fix(repo) something" + additional,
			want: commit.NewGroup("Fix", "repo", "something"+additional, false),
		},
		"topic section colon": {
			line: "Fix(repo): something",
			want: commit.NewGroup("Fix", "repo", "something", false),
		},
		"topic section colon multi": {
			line: "Fix(repo): something" + additional,
			want: commit.NewGroup("Fix", "repo", "something"+additional, false),
		},
		"ref":          {line: "ref something", want: commit.NewGroup("Refactor", "", "something", false)},
		"refactor":     {line: "refactor something", want: commit.NewGroup("Refactor", "", "something", false)},
		"feat":         {line: "feat something", want: commit.NewGroup("Feature", "", "something", false)},
		"feature":      {line: "feature something", want: commit.NewGroup("Feature", "", "something", false)},
		"fix":          {line: "fix something", want: commit.NewGroup("Fix", "", "something", false)},
		"fixed":        {line: "fixed something", want: commit.NewGroup("Fix", "", "something", false)},
		"chore":        {line: "chore something", want: commit.NewGroup("Chore", "", "something", false)},
		"upgrade":      {line: "upgrade something", want: commit.NewGroup("Upgrades", "", "something", false)},
		"enhance":      {line: "enhance something", want: commit.NewGroup("Enhancements", "", "something", false)},
		"enhancement":  {line: "enhancement something", want: commit.NewGroup("Enhancements", "", "something", false)},
		"enhancements": {line: "enhancements something", want: commit.NewGroup("Enhancements", "", "something", false)},
		"style":        {line: "style something", want: commit.NewGroup("Style", "", "something", false)},
		"comma sep":    {line: "fix(git,commit): something", want: commit.NewGroup("Fix", "git,commit", "something", false)},
		"hyphen subj":  {line: "fix(git-commit): something", want: commit.NewGroup("Fix", "git-commit", "something", false)},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := commit.GroupFromCommit(tc.line)
			if diff := cmp.Diff(tc.want, got, commit.GroupComparer...); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestGroup(t *testing.T) {
	t.Parallel()
	t.Run("DescriptionString", testGroupDescriptionString)
	t.Run("ParseGroups", testGroupParseGroups)
}

func testGroupDescriptionString(t *testing.T) {
	t.Parallel()
	// the first letter will be uppercased.
	letter := testament.RandomLowerString(1)
	randomString := testament.RandomLowerString(30)
	msg := letter + randomString
	wantMsg := strings.ToUpper(letter) + randomString
	additional := `\n\n` + testament.RandomString(50)
	prefix := commit.ItemPrefix
	issue := "Close #666"

	tcs := map[string]struct {
		group commit.Group
		want  string
	}{
		"simple": {
			group: commit.NewGroup("Fix", "", msg, false),
			want:  prefix + wantMsg,
		},
		"with verb": {
			group: commit.NewGroup("Fix", "repo", msg, false),
			want:  fmt.Sprintf("%s**Repo:** %s", prefix, wantMsg),
		},
		"with ci verb": {
			group: commit.NewGroup("Fix", "ci", msg, false),
			want:  fmt.Sprintf("%s**CI:** %s", prefix, wantMsg),
		},
		"multi line": {
			group: commit.NewGroup("Fix", "repo", msg+additional, false),
			want:  fmt.Sprintf("%s**Repo:** %s", prefix, wantMsg),
		},
		"with issue ref": {
			group: commit.NewGroup("Fix", "repo", msg+additional+`\n`+issue, false),
			want:  fmt.Sprintf("%s**Repo:** %s (%s)", prefix, wantMsg, issue),
		},
		"with issue (ref)": {
			group: commit.NewGroup("Fix", "repo", msg+additional+`\n(`+issue+`)`, false),
			want:  fmt.Sprintf("%s**Repo:** %s (%s)", prefix, wantMsg, issue),
		},
		"multi issue refs": {
			group: commit.NewGroup("Fix", "repo", msg+additional+`\n`+issue+`\n`+issue, false),
			want:  fmt.Sprintf("%s**Repo:** %s (%s, %s)", prefix, wantMsg, issue, issue),
		},
		"comma separated": {
			group: commit.NewGroup("Fix", "git,commit", msg, false),
			want:  fmt.Sprintf("%s**Git,Commit:** %s", prefix, wantMsg),
		},
		"hyphenated subjects": {
			group: commit.NewGroup("Fix", "git-commit", msg, false),
			want:  fmt.Sprintf("%s**Git-commit:** %s", prefix, wantMsg),
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got := tc.group.DescriptionString()
			if diff := cmp.Diff(tc.want, got, commit.GroupComparer...); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func testGroupParseGroups(t *testing.T) {
	t.Run("OneGroup", testGroupParseGroupsOneGroup)
	t.Run("MultipleGroups", testGroupParseGroupsMultipleGroups)
	t.Run("BreakingSign", testGroupParseGroupsBreakingSign)
	t.Run("BreakingFooter", testGroupParseGroupsBreakingFooter)
}

func testGroupParseGroupsOneGroup(t *testing.T) {
	t.Parallel()
	logs := []string{"Feat(testing): this is a test"}
	got := commit.ParseGroups(logs)

	got = strings.TrimRight(got, "\n")
	want := "### Feature\n\n- **Testing:** This is a test"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func testGroupParseGroupsMultipleGroups(t *testing.T) {
	t.Parallel()
	logs := []string{
		"Feat(testing): this is a test",
		"Misc: this is another test",
		"feat: yet another",
	}
	got := commit.ParseGroups(logs)

	want := []string{
		"### Feature\n\n- **Testing:** This is a test\n- Yet another",
		"### Misc\n\n- This is another test",
	}
	gotS := strings.Split(got, "\n\n\n")
	sort.Strings(gotS)
	if diff := cmp.Diff(want, gotS); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func testGroupParseGroupsBreakingSign(t *testing.T) {
	t.Parallel()
	logs := []string{
		"ref: nothing important",
		"ref!(repo): this is a test",
	}
	got := commit.ParseGroups(logs)

	want := strings.Join([]string{
		"### Refactor\n",
		"- Nothing important",
		"- **Repo:** This is a test [**BREAKING CHANGE**]",
	}, "\n")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func testGroupParseGroupsBreakingFooter(t *testing.T) {
	t.Parallel()
	logs := []string{
		"ref(server): nothing special",
		"ref(repo): this is a new api\n\nBREAKING CHANGE: this is a changed api",
	}
	got := commit.ParseGroups(logs)

	want := strings.Join([]string{
		"### Refactor\n",
		"- **Server:** Nothing special",
		"- **Repo:** This is a new api [**BREAKING CHANGE**]",
	}, "\n")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}
