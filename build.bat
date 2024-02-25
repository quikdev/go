@echo off
echo building...
cd ./src
go mod tidy
go build -ldflags "-X 'main.name=qgo' -X 'main.description=A Go development utility' -X 'main.version=1.0.0'" -o ..\bin\qgo.exe qgo.go
cd ..\
echo build complete
@echo on
