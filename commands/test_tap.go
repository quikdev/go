package commands

import (
	"fmt"
	"strings"

	tap "github.com/mpontillo/tap13"
)

type Formatter interface {
	Format(results *tap.Results)
	Summary()
}

type TapOutput struct {
	Passed     bool
	TestNumber int
	Test       string
	Label      string
	Diagnostic []string
}

func (o *TapOutput) Format() string {
	success := ""
	if o.Passed == false {
		success = "not"
	}

	txt := strings.TrimSpace(fmt.Sprintf("%v ok %d - %v\n", success, o.TestNumber, o.Test))

	if o.Label != emptyString {
		txt = fmt.Sprintf("# %v\n%v\n", o.Label, txt)
	}

	if len(o.Diagnostic) > 0 {
		txt += fmt.Sprintf("  ---\n  message: \"%v\"\n  ...\n", strings.ReplaceAll(strings.Join(o.Diagnostic, "\n  "), "\"", "'"))
	}

	return txt
}
