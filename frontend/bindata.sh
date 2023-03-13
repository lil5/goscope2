#!/bin/sh

go install github.com/kevinburke/go-bindata/...@latest

npm run build

go-bindata -pkg frontend dist/