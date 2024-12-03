set set GO111MODULE=on
@REM set GOROOT=d:\SyncThreeUnzip\go_1_15_15
set GOROOT=d:\SyncThreeUnzip\go_1_21_4
set PATH=%GOROOT%\bin;%PATH%

@REM GOOS=windows go build -ldflags -H=windowsgui .
go build -ldflags -H=windowsgui .