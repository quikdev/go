package main

import "testing"

func TestSayHello(t *testing.T) {
  want := "Name: {{ .Name }}\nVersion: {{ .Version }}\nDescription: {{ .Description }}\n"
  got := SayHello()

  if got != expected {
    t.Errorf("SayHello() = %q, want %q", got, want)
  }
}
