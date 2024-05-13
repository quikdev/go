package context

import (
	"fmt"
	"strings"

	"github.com/quikdev/go/util"
)

type args struct {
	flag string
	data []string
}

func NewArgs(flag string) *args {
	return &args{flag: flag, data: []string{}}
}

func (a *args) Add(key string, value ...string) {
	val := key
	if len(value) > 0 {
		val += "=" + value[0]
	}

	if !util.InSlice[string](val, a.data) {
		a.data = append(a.data, val)
	}
}

func (a *args) Size() int {
	return len(a.data)
}

func (a *args) String(quoted ...bool) string {
	quote := false
	if len(quoted) > 0 {
		quote = quoted[0]
	}

	q := ""
	if quote {
		q = `"`
	}

	return fmt.Sprintf(`-%s=%s%s%s`, a.flag, q, strings.Join(a.data, " "), q)
}
