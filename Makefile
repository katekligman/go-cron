TOKEN = `cat .token`
REPO := go-cron
USER := odise
VERSION := "v0.0.4"

build:
	mkdir -p out/linux
	GOOS=linux go build -o out/linux/go-cron -ldflags "-X main.build `git rev-parse --short HEAD`" bin/go-cron.go

release: build
	rm -f out/linux/go-cron-linux.gz
	gzip -c out/linux/go-cron > out/linux/go-cron-linux.gz

