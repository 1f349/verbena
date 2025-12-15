.PHONY: all sqlc build deb

all: sqlc

sqlc:
	sqlc generate

build:
	mkdir -p dist/
	go build -trimpath -ldflags "-s -w" -o ./dist/verbena ./cmd/verbena

deb: build
	mkdir -p packaging/debian/usr/bin/
	cp dist/verbena packaging/debian/usr/bin/verbena
	cd packaging/debian/ && dpkg-buildpackage -us -uc -ui -b
