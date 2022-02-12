package commit

// NewGroup returns a new instance of the Group.
func NewGroup(sec, subject, desc string) Group {
	return Group{
		Verb:        sec,
		Subject:     subject,
		Description: desc,
	}
}
