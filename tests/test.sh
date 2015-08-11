#!/usr/bin/env sh

mkdir --parents ~/.eris/languages
mv config.json ~/.eris/languages/.
go test ../...