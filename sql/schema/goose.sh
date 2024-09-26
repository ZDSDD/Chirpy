#!/bin/bash

case "$1" in
    "up")
        goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" up
        ;;
    "down")
        goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" down
        ;;
    *)
        echo "Invalid argument"
        ;;
    esac    