#!/usr/bin/env bash

hyperfine -p "{runner} && dagger core engine local-cache prune" "{runner} && dagger call --mod mod28 fn" -L runner "unset _EXPERIMENTAL_DAGGER_RUNNER_HOST,export _EXPERIMENTAL_DAGGER_RUNNER_HOST=\"tcp://$APP_NAME.internal:2345\""
