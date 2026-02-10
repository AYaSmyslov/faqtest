#!/bin/bash

swag init \
    -g main.go \
    -d ./cmd/api,./internal \
    -o ./docs \
    --parseInternal
