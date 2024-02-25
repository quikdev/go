package {{ .Module }}

import "testing"

func TestHello(t *testing.T) {
  name = "{{ .Name }}"
  description = "{{ .Description }}"
  version = "{{ .Version }}"

  person := "Alice"
  want := "Hello, Alice!\n{{ .Name }} v{{ .Version }}\n{{ .Description }}"
  if got := Hello(person); got != want {
    t.Errorf("Hello() = %q, want %q", got, want)
  }
}
