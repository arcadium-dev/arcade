#!/bin/bash

# shellcheck disable=2034

if [[ "${BASH_COMMON_SH:-}" != "" ]]; then return; fi
declare -r BASH_COMMON_SH="true"

[[ "${DEBUG:-}" != "" ]] && set -o xtrace

set -o errexit
set -o nounset
set -o pipefail
set -o noglob

IFS=$'\t\n'

declare -ri SUCCESS=0
declare -ri FAILURE=1

#
# Logging
#

declare -r ESC=$'\033'
declare -r CLEAR="${ESC}[0;39m"

declare -r RED="${ESC}[1;31m"
declare -r GREEN="${ESC}[1;32m"
declare -r YELLOW="${ESC}[1;33m"
declare -r BLUE="${ESC}[1;34m"
declare -r MAGENTA="${ESC}[1;35m"
if [[ "${DEBUG:-}" != "" ]]; then
  echo -en "${CLEAR}"
fi

#----------------------------------------------------------------------------
# backtrace
#----------------------------------------------------------------------------
function backtrace() {
  local i
  for (( i = 0; i < $(( ${#FUNCNAME[@]} - 1 )); i++ )); do
    printf "at %20s, %20s, line %s\n" \
      "${FUNCNAME[$i]}" \
      "${BASH_SOURCE[$((i + 1))]}" \
      "${BASH_LINENO[${i}]}"
  done
}

#----------------------------------------------------------------------------
# panic ...
#----------------------------------------------------------------------------
function panic() {
  error "$*"
  backtrace
  exit ${FAILURE}
}

#----------------------------------------------------------------------------
# error ...
#----------------------------------------------------------------------------
function error() {(
  IFS=$' '
  local message="$*"
  param_check "${message}"
  echo -e "${RED}Error:${CLEAR} ${message}"
)}

#----------------------------------------------------------------------------
# die ...
#----------------------------------------------------------------------------
function die() {
  error "$*"
  exit ${FAILURE}
}

#----------------------------------------------------------------------------
# warn ...
#----------------------------------------------------------------------------
function warn() {(
  IFS=$' '
  local message="$*"
  param_check "${message}"
  echo -e "${MAGENTA}Warning:${CLEAR} ${message}"
)}

#----------------------------------------------------------------------------
# info
#----------------------------------------------------------------------------
function info() {(
  IFS=$' '
  if [[ $# -eq 2 ]]; then
    local heading="${1:-}" body="${2:-}"
    param_check "${heading}" "${body}"
    echo -e "\n${YELLOW}${heading}:${CLEAR} ${body}"
  else
    local message="$*"
    param_check "${message}"
    echo -e "\n${YELLOW}${message}${CLEAR}"
  fi
)}

#----------------------------------------------------------------------------
# detail
#----------------------------------------------------------------------------
function detail() {(
  IFS=$' '
  if [[ $# -eq 2 ]]; then
    local heading="${1:-}" body="${2:-}"
    param_check "${heading}" "${body}"
    echo -e "${BLUE}${heading}:${CLEAR} ${body}"
  else
    local message="$*"
    param_check "${message}"
    echo -e "${BLUE}${message}${CLEAR}"
  fi
)}

#----------------------------------------------------------------------------
# msg
#----------------------------------------------------------------------------
function msg() {(
  IFS=$' '
  if [[ $# -eq 0 ]]; then
    local line
    while read -r line; do
      echo -e "${indent:-}${line}"
    done
  else
    local message="$*"
    param_check "${message}"
    echo -e "${indent:-}${message}"
  fi
)}

#----------------------------------------------------------------------------
# success
#----------------------------------------------------------------------------
function success() {
  echo -e "${GREEN}Success${CLEAR}"
  return "${SUCCESS}"
}

#----------------------------------------------------------------------------
# failed
#----------------------------------------------------------------------------
function failed() {
  echo -e "${RED}Failed${CLEAR}"
  exit "${FAILURE}"
}

#----------------------------------------------------------------------------
# param_check ...
#----------------------------------------------------------------------------
function param_check() {
  local param
  for param in "$@"; do
    if [[ "${param}" == "" ]]; then
      panic "Missing parameter"
    fi
  done
}

#----------------------------------------------------------------------------
# cmd_check ...
#----------------------------------------------------------------------------
function cmd_check() {
  local cmd
  for cmd in "$@"; do
    if ! type "${cmd}" &>/dev/null; then
      return ${FAILURE}
    fi
  done
  return ${SUCCESS}
}
