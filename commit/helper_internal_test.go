package commit

import (
	"github.com/google/go-cmp/cmp"
)

var GroupComparer = []cmp.Option{
	cmp.AllowUnexported(Group{}),
	cmp.Transformer("GroupFixer", func(in Group) Group {
		in.raw = ""
		return in
	}),
}
