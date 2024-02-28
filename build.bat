@echo off
echo building...
go mod tidy
cd .\cmd\qgo
go build -ldflags "-X 'main.name=qgo' -X 'main.description=Quickly create, run, and test Go with QuikGo' -X 'main.version=1'" -o ..\..\bin\qgo.exe qgo.go
cd ..\..\
echo build complete
@echo on
