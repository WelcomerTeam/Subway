#!/bin/bash

gci write --skip-generated -s default subway
gofumpt -d -e -extra -l -w subway
go mod tidy