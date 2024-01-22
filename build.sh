mkdir temp
mkdir dist
GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -v -ldflags="-w -s" -o ./dist/linux-amd64/speedtest main.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=0  go build -v -ldflags="-w -s" -o ./dist/windows-amd64/speedtest.exe main.go 
CC=mips-linux-gcc GOOS=linux GOARCH=mipsle CGO_ENABLED=0 GOMIPS=softfloat go build -v -ldflags="-w -s" -o ./dist/linux-mipsle-softfloat/speedtest main.go 
GOOS=linux GOARCH=arm64 CGO_ENABLED=0  go build -v -ldflags="-w -s" -o ./dist/linux-arm64/speedtest main.go  
wget --output-document=./temp/speedtest-go-main.zip https://github.com/masx200/speedtest-go/archive/refs/heads/main.zip
deno run -A ./build.ts