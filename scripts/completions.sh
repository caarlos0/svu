#!/bin/sh
set -e
rm -rf completions
mkdir completions
go build -o svu .
for sh in bash zsh fish; do
  ./svu completion "$sh" >"completions/svu.$sh"
done
