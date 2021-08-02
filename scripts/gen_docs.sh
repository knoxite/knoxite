#!/bin/sh
set -e

rm -rf docs/knoxite*
go run -tags docs ./cmd/knoxite docs ./docs
