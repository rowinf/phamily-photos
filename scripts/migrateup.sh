#!/bin/bash

if [ -f .env ]; then
    source .env
fi
go build -o goose-custom ./cmd/migrate/*.go
./goose-custom
rm ./goose-custom
