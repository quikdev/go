// Import this into your project using
// import "{{ .Module }}/internal/{{ .ModuleName }}"
package {{ .ModuleName }}

import "fmt"

func {{ .FuncName }}(name string) string {
  return fmt.Sprintf("{{ .Greeting }}, %v!", name)
}
