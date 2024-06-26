{
  "name": "{{ .Name }}",
  "version": "0.0.1",
  "description": "{{ .Description }}",
  "license": "{{ .License }}",
  "author": "{{ .Author }}",
  "build": "{{ .Main }}",
  "update": true,
  "upx": false,
  "tiny": false,
  "variables": {
    "main.name": "manifest.name",
    "main.version": "manifest.version",
    "main.description": "manifest.description"
  },
  "minify": true
}