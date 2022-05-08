package commit

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	descRe = regexp.MustCompile(`^\s*([[:alpha:]]+!?)\(?([[:alpha:],_-]+)?\)?(!)?:?(.*)`)
	refRe  = regexp.MustCompile(`[[:alpha:]]+\s+#\d+`)
)

// ItemPrefix is the markdown prefix before each item.
var ItemPrefix = "- "

// A Group is a commit with all of its messages.
type Group struct {
	raw         string
	Verb        string
	Subject     string
	Description string
	Breaking    bool
}

// GroupFromCommit creates a Group object from the given line.
func GroupFromCommit(msg string) Group {
	matches := descRe.FindStringSubmatch(msg)
	verb := matches[1]
	subject := matches[2]
	verbBreak := matches[3]
	desc := matches[4]

	breaking := false
	if strings.HasSuffix(verb, "!") || verbBreak != "" {
		breaking = true
		verb = strings.TrimSuffix(verb, "!")
	}

	switch strings.ToLower(verb) {
	case "ref", "refactor":
		verb = "Refactor"
	case "feat", "feature":
		verb = "Feature"
	case "fix", "fixed":
		verb = "Fix"
	case "chore":
		verb = "Chore"
	case "enhance", "enhancements", "enhancement":
		verb = "Enhancements"
	case "upgrade":
		verb = "Upgrades"
	case "ci":
		verb = "CI"
	case "style":
		verb = "Style"
	case "docs":
		verb = "Docs"
	default:
		verb = ""
	}

	if verb == "" {
		verb = "Misc"
	}
	if desc == "" {
		desc = matches[0]
	}

	return Group{
		raw:         msg,
		Verb:        verb,
		Subject:     subject,
		Description: strings.TrimSpace(desc),
		Breaking:    breaking,
	}
}

// Section returns a printable line for the section.
func (g Group) Section() string {
	return "### " + upperFirst(g.Verb)
}

// DescriptionString returns a string that is suitable for printing a line in a
// Group.
func (g Group) DescriptionString() string {
	subject := g.Subject
	if strings.EqualFold(subject, "ci") {
		subject = "CI"
	}
	if subject != "" {
		subjects := strings.Split(subject, ",")
		for i := range subjects {
			subjects[i] = upperFirst(subjects[i])
		}
		subject = strings.Join(subjects, ",")
		subject = "**" + subject + ":** "
	}

	lines := strings.Split(g.Description, `\n`)
	refs := make([]string, 0, len(lines))
	title := lines[0]
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		matches := refRe.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			refs = append(refs, match...)
		}
	}

	title = strings.TrimPrefix(title, " ")
	var ref string
	if len(refs) > 0 {
		ref = fmt.Sprintf(" (%s)", strings.Join(refs, ", "))
	}
	return fmt.Sprintf("- %s%s%s", subject, upperFirst(title), ref)
}

// ParseGroups parses the lines in the logs and returns them as a string.
func ParseGroups(logs []string) string {
	logs = cleanup(logs)
	groups := make(map[string][]Group, len(logs))
	for _, line := range logs {
		group := GroupFromCommit(line)
		groups[group.Verb] = append(groups[group.Verb], group)
	}

	buf := &strings.Builder{}
	i := 0
	for _, desc := range groups {
		fmt.Fprintln(buf, desc[0].Section()+"\n")
		for _, line := range desc {
			fmt.Fprint(buf, line.DescriptionString())
			if line.Breaking {
				fmt.Fprintf(buf, " [**BREAKING CHANGE**]")
			}
			fmt.Fprintln(buf, "")
		}
		i++
		if i < len(groups) {
			fmt.Fprintf(buf, "\n\n")
		}
	}

	str := buf.String()
	return strings.TrimSuffix(str, "\n")
}

// cleanup returns only the title of the logs.
func cleanup(logs []string) []string {
	ret := make([]string, 0, len(logs))
	for _, commit := range logs {
		items := strings.Split(commit, "\n")
		item := items[0]
		breaking := false
		for _, line := range items[1:] {
			if strings.Contains(line, "BREAKING CHANGE") {
				breaking = true
			}
			if !strings.Contains(line, "#") {
				continue
			}
			item = fmt.Sprintf("%s (%s)", item, line)
		}
		if breaking {
			item += " [**BREAKING CHANGE**]"
		}
		if item == "" {
			continue
		}
		item = strings.TrimPrefix(item, " ")
		ret = append(ret, item)
	}
	return ret
}

// upperFirst makes the first letter of the string an uppercase letter.
func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
