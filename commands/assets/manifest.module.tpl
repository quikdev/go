{
  "name": "{{ .Name }}",
  "version": "0.0.1",
  "description": "{{ .Description }}",
  "license": "{{ .License }}",
  "author": "{{ .Author }}",
  "variables": {
    "{{ .Module }}.name": "manifest.name",
    "{{ .Module }}.version": "manifest.version",
    "{{ .Module }}.description": "manifest.description"
  },
  "test": {
    "format": "spec",
    "debug":  true
  }
}