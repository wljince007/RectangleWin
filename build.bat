set set GO111MODULE=on
@REM set GOROOT=d:\SyncThreeUnzip\go1.15.15
set GOROOT=d:\SyncThreeUnzip\go1.21.4
set PATH=%GOROOT%\bin;%PATH%

@REM GOOS=windows go build -ldflags -H=windowsgui .
go build -ldflags -H=windowsgui .