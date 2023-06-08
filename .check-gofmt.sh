#!/bin/bash

# exclude vendor/ ; see https://stackoverflow.com/a/4210072
SOURCES=$(find . -path ./vendor -prune -o -name '*.go' -print)

if [ "$1" = "--fix" ]; then
	exec gofmt -s -w $SOURCES
fi

if [ -n "$(gofmt -s -l $SOURCES)" ]; then
    echo "Go code is not formatted correctly with \`gofmt -s\`:"
    gofmt -s -d $SOURCES
    exit 1
fi
