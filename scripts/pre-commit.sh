#!/bin/sh

# Run packaged gookme pre-commit hook
go run ./cmd/cli r -t pre-commit $1
