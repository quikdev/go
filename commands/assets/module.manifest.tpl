{
  "name": "{{ .Name }}",
  "version": "{{ .Version }}",
  "description": "{{ .Description }}",
  "license": "{{ .License }}",
  "author": "{{ .Author }}",
  "build": "{{ .Main }}",
  "variables": {
    "name": "manifest.name",
    "version": "manifest.version",
    "description": "manifest.description"
  },
  "minify": true,
  "test": {
    "format": "spec",
    "debug":  true
  }
}