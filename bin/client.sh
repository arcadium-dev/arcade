#!/usr/bin/env bash

declare -i _failures

title() {
  _failures=0
  IFS=$' '
  local message="$*"
  param_check "${message}"
  echo -e "\n${MAGENTA}Test:${CLEAR} ${message}"
}

pass() {
  IFS=$' '
  local message="$*"
  param_check "${message}"
  echo -e "${BLUE}Test passed:${CLEAR} ${message}"
}

fail() {
  (( _failures++ ))
  error "$@"
}

fatal() {
  fail "$@"
  failed
}

report() {
  if [[ $_failures -ne 0 ]]; then
    echo -e "${RED}Test failures:${CLEAR} ${_failures}"
    failed
  fi
  success
}

data_field() {
  local field="$1" body="$2"
  param_check "${field}" "${body}"

  jq -r ".data.${field}" <(echo "${body}")
}

error_detail() {
  error_field "detail" "$1"
}

error_status() {
  error_field "status" "$1"
}

error_field() {
  local field="$1" body="$2"
  param_check "${field}" "${body}"

  jq -r ".error.${field}" <(echo "${body}")
}

is_error() {
  local err
  err="$(jq .error <(echo "${@:-}"))" 
  if [[ "${err}" == "null" || "${err}" == "" ]]; then
    return ${FAILURE}
  fi
  return ${SUCCESS}
}

skip_broken_on_ci() {
  if [[ "${CI:-}" == ""  ]]; then
    return
  fi
  skip_broken
}

declare -ri SKIPPED=2

skip_broken() {
  echo -e "\n${BLUE}Skipping${CLEAR}"
  exit ${SKIPPED}
}

contains() {
  local sub="$1" s="$2"
  if [[ "${s/${sub}}" == "${s}" ]]; then
    return ${FAILURE}
  fi
  return ${SUCCESS}
}

players_list() {
  info "Listing players" >&2
  
  local result
  result="$(bin/dev run curl --request GET "https://assets:4201/players" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

players_get() {
  local _id="$1"
  param_check "${_id}"

  info "Getting player" >&2
  msg "id: ${_id}" >&2

  local result
  result="$(bin/dev run curl --request GET "https://assets:4201/players/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

players_create() {
  local _name="$1" _desc="$2" _home="$3" _location="$4"
  param_check "${_name}" "${_desc}" "${_home}" "${_location}"

  info "Creating player" >&2
  msg "name:        ${_name}" >&2
  msg "description: ${_desc}" >&2
  msg "home:        ${_home}" >&2
  msg "location:    ${_location}" >&2

  local result
  result="$(bin/dev run curl --request POST --data '{"name":"'"${_name}"'","description":"'"${_desc}"'","home":"'"${_home}"'","location":"'"${_location}"'"}' "https://assets:4201/players" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

players_update() {
  local _id="$1" _name="$2" _desc="$3" _home="$4" _location="$5"
  param_check "${_id}" "${_name}" "${_desc}" "${_home}" "${_location}"

  info "Updating player" >&2
  msg "id:          ${_id}" >&2
  msg "name:        ${_name}" >&2
  msg "description: ${_desc}" >&2
  msg "home:        ${_home}" >&2
  msg "location:    ${_location}" >&2

  local result
  result="$(bin/dev run curl --request PUT --data '{"playerID":"'"${_id}"'","name":"'"${_name}"'","description":"'"${_desc}"'","home":"'"${_home}"'","location":"'"${_location}"'"}' "https://assets:4201/players/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

players_remove() {
  local _id="$1"
  param_check "${_id}"

  info "Removing player" >&2
  msg "player id: ${_id}" >&2

  local result
  result="$(bin/dev run curl --request DELETE "https://assets:4201/players/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

rooms_get() {
  local _id="$1"
  param_check "${_id}"

  info "Getting room" >&2
  msg "id: ${_id}" >&2

  local result
  result="$(bin/dev run curl --request GET "https://assets:4201/rooms/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

rooms_create() {
  local _name="$1" _desc="$2" _owner="$3" _parent="$4"
  param_check "${_name}" "${_desc}" "${_owner}" "${_parent}"

  info "Creating room" >&2
  msg "name:        ${_name}" >&2
  msg "description: ${_desc}" >&2
  msg "owner:       ${_owner}" >&2
  msg "parent:      ${_parent}" >&2

  local result
  result="$(bin/dev run curl --request POST --data '{"name":"'"${_name}"'","description":"'"${_desc}"'","owner":"'"${_owner}"'","parent":"'"${_parent}"'"}' "https://assets:4201/rooms" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

rooms_update() {
  local _id="$1" _name="$2" _desc="$3" _owner="$4" _parent="$5"
  param_check "${_id}" "${_name}" "${_desc}" "${_owner}" "${_parent}"

  info "Updating room" >&2
  msg "id:          ${_id}" >&2
  msg "name:        ${_name}" >&2
  msg "description: ${_desc}" >&2
  msg "owner:       ${_owner}" >&2
  msg "parent:      ${_parent}" >&2

  local result
  result="$(bin/dev run curl --request PUT --data '{"roomID":"'"${_id}"'","name":"'"${_name}"'","description":"'"${_desc}"'","owner":"'"${_owner}"'","parent":"'"${_parent}"'"}' "https://assets:4201/rooms/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

rooms_remove() {
  local _id="$1"
  param_check "${_id}"

  info "Removing room" >&2
  msg "room id: ${_id}" >&2

  local result
  result="$(bin/dev run curl --request DELETE "https://assets:4201/rooms/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

links_get() {
  local _id="$1"
  param_check "${_id}"

  info "Getting link" >&2
  msg "id: ${_id}" >&2

  local result
  result="$(bin/dev run curl --request GET "https://assets:4201/links/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

links_create() {
  local _name="$1" _desc="$2" _owner="$3" _location="$4" _destination="$5"
  param_check "${_name}" "${_desc}" "${_owner}" "${_location}" "${_destination}"

  info "Creating link" >&2
  msg "name:        ${_name}" >&2
  msg "description: ${_desc}" >&2
  msg "owner:       ${_owner}" >&2
  msg "location:    ${_location}" >&2
  msg "destination: ${_destination}" >&2

  local result
  result="$(bin/dev run curl --request POST --data '{"name":"'"${_name}"'","description":"'"${_desc}"'","owner":"'"${_owner}"'","location":"'"${_location}"'","destination":"'"${_destination}"'"}' "https://assets:4201/links" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

links_update() {
  local _id="$1" _name="$2" _desc="$3" _owner="$4" _location="$5" _destination="$6"
  param_check "${_id}" "${_name}" "${_desc}" "${_owner}" "${_location}" "${_destination}"

  info "Updating link" >&2
  msg "id:          ${_id}" >&2
  msg "name:        ${_name}" >&2
  msg "description: ${_desc}" >&2
  msg "owner:       ${_owner}" >&2
  msg "location:    ${_location}" >&2
  msg "destination: ${_destination}" >&2

  local result
  result="$(bin/dev run curl --request PUT --data '{"linkID":"'"${_id}"'","name":"'"${_name}"'","description":"'"${_desc}"'","owner":"'"${_owner}"'","location":"'"${_location}"'","destination":"'"${_destination}"'"}' "https://assets:4201/links/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

links_remove() {
  local _id="$1"
  param_check "${_id}"

  info "Removing link" >&2
  msg "link id: ${_id}" >&2

  local result
  result="$(bin/dev run curl --request DELETE "https://assets:4201/links/${_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result:-}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}
