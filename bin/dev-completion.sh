#!/usr/bin/env bash

_dev() {
	local cmd
	COMPREPLY=()
	for cmd in $("${1}" completion "${3}"); do
		if [[ "${cmd}" == "${2}"* ]]; then
			COMPREPLY+=("${cmd}")
		fi
	done
}
complete -F _dev -o default dev
