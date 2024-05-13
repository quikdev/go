package {{ .Module }}

import "fmt"

var (
  name        string
  description string
  version     string
)

// TODO: Write my code
// Hello returns a greeting for the named person.
func Hello(person string) string {
  return fmt.Sprintf("Hello, %v!\n%v v%v\n%v", person, name, version, description)
}
