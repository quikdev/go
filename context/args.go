package context

import (
	"strings"

	"github.com/quikdev/qgo/v1/util"
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

func (a *args) String() string {
	return "-" + a.flag + " \"" + strings.Join(a.data, " ") + "\""
}
