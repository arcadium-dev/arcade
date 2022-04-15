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

  jq -r ".errors[0].${field}" <(echo "${body}")
}

is_error() {
  local err
  err="$(jq .errors[0] <(echo "${@:-}"))" 
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

deployments_create() {
  local _orchestrator_id="$1" _token="$2" _domain_name="$3"
  param_check "${_orchestrator_id}" "${_token}" "${_domain_name}"

  info "Creating deployment" >&2
  msg "orchestrator id:   ${_orchestrator_id}" >&2
  msg "regstration token: ${_token}" >&2
  msg "domain name:       ${_domain_name}" >&2

  local result
  result="$(bin/dev run curl --request POST --data '{"orchestratorID":"'${_orchestrator_id}'", "registrationToken": "'${_token}'", "domainName": "'${_domain_name}'"}' "https://infra:8443/deployments" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

deployments_list() {
  info "Listing deployments" >&2
  
  local result
  result="$(bin/dev run curl --request GET "https://infra:8443/deployments" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

deployments_get() {
  local _orchestrator_id="$1"
  param_check "${_orchestrator_id}"

  info "Getting deployment" >&2
  msg "orchestrator id: ${_orchestrator_id}" >&2

  local result
  result="$(bin/dev run curl --request GET "https://infra:8443/deployments/${_orchestrator_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

deployments_remove() {
  local _orchestrator_id="$1"
  param_check "${_orchestrator_id}"

  info "Removing deployment" >&2
  msg "orchestrator id: ${_orchestrator_id}" >&2

  local result
  result="$(bin/dev run curl --request DELETE "https://infra:8443/deployments/${_orchestrator_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

hostnames_create() {
  local _orchestrator_id="$1" _hostname="$2"
  param_check "${_orchestrator_id}" "${_hostname}"

  info "Creating hostname" >&2
  msg "orchestrator id: ${_orchestrator_id}" >&2
  msg "hostname:        ${_hostname}" >&2

  local result
  result="$(bin/dev run curl --request POST --data '{"hostname": "'${_hostname}'"}' "https://infra:8443/deployments/${_orchestrator_id}/hostnames" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

hostnames_list() {
  local _orchestrator_id="$1"
  param_check "${_orchestrator_id}"

  info "Listing hostname" >&2
  msg "orchestrator id: ${_orchestrator_id}" >&2

  local result
  result="$(bin/dev run curl --request GET  "https://infra:8443/deployments/${_orchestrator_id}/hostnames" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

hostnames_get() {
  local _orchestrator_id="$1" _hostname_id="$2"
  param_check "${_orchestrator_id}" "${_hostname_id}"

  info "Geting hostname" >&2
  msg "orchestrator id: ${_orchestrator_id}" >&2
  msg "hostname_id:     ${_hostname_id}" >&2

  local result
  result="$(bin/dev run curl --request GET "https://infra:8443/deployments/${_orchestrator_id}/hostnames/${_hostname_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

hostnames_status() {
  local _orchestrator_id="$1" _hostname_id="$2"
  param_check "${_orchestrator_id}" "${_hostname_id}"

  info "Updating hostname status" >&2
  msg "orchestrator id: ${_orchestrator_id}" >&2
  msg "hostname_id:     ${_hostname_id}" >&2

  local result
  result="$(bin/dev run curl --request GET "https://infra:8443/deployments/${_orchestrator_id}/hostnames/${_hostname_id}/status" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}

hostnames_remove() {
  local _orchestrator_id="$1" _hostname_id="$2"
  param_check "${_orchestrator_id}" "${_hostname_id}"

  info "Removing hostname" >&2
  msg "orchestrator id: ${_orchestrator_id}" >&2
  msg "hostname_id:     ${_hostname_id}" >&2

  local result
  result="$(bin/dev run curl --request DELETE "https://infra:8443/deployments/${_orchestrator_id}/hostnames/${_hostname_id}" 2>/dev/null)"
  local rc=$?

  if [[ "${result}" != "" ]]; then
    msg "\nResponse\n$(jq . <(echo "${result}"))" >&2
    echo "${result}" >&1
  fi
  return ${rc}
}
