#!/bin/sh

# Run packaged gookme commit-msg hook
go run ./cmd/cli r -t commit-msg $1
