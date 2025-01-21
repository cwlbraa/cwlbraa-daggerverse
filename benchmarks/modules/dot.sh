#!/usr/bin/env bash

set -ex

dagger core engine local-cache prune
dagger --dot-output ~/mods.dot call --mod mod28 fn
dot -Tsvg -x -o ~/mods-out.svg ~/mods.dot
open -a Safari ~/mods-out.svg
