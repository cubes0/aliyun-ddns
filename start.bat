@REM windows平台编译运行
@REM SET GOPROXY=https://goproxy.io,direct
go mod tidy && go build -o aliyun-ddns.exe main.go
start aliyun-ddns.exe