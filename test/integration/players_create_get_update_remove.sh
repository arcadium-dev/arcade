test_players_create_get_update_delete() {
  title "Players: create, get, update and delete"

  local name="$(tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)))"
  local desc="$(for i in {1..10}; do tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)); echo -n " "; done)"
  local home="00000000-0000-0000-0000-000000000001"
  local location="00000000-0000-0000-0000-000000000001"
  local resp id actual

  # Create a player
  resp="$(players_create "${name}" "${desc}" "${home}" "${location}")"
  if is_error "${resp}"; then
    fatal "$(error_detail "${resp}")"
  fi
  id="$(data_field "playerID" "${resp}")"


  # Get the player
  resp="$(players_get "${id}")"
  if is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  # Check that the returned name, description, home and location
  actual="$(data_field "name" "${resp}")"
  if [[ "${actual}" != "${name}" ]]; then
    fail "Expected name ${name}, actual ${actual}"
  else
    pass "name matches"
  fi

  actual="$(data_field "description" "${resp}")"
  if [[ "${actual}" != "${desc}" ]]; then
    fail "Expected description '${token}', actual '${actual}'"
  else
    pass "description matches"
  fi

  actual="$(data_field "home" "${resp}")"
  if [[ "${actual}" != "${home}" ]]; then
    fail "Expected home ${home}, actual ${actual}"
  else
    pass "home matches"
  fi

  actual="$(data_field "location" "${resp}")"
  if [[ "${actual}" != "${location}" ]]; then
    fail "Expected location ${location}, actual ${actual}"
  else
    pass "location matches"
  fi

  # Update the player name
  name="$(tr -cd 'A-Za-z' < /dev/urandom | head -c $(( $RANDOM % 7  + 2)))"
  resp="$(players_update "${id}" "${name}" "${desc}" "${home}" "${location}")"
  if is_error "${resp}"; then
    fatal "$(error_detail "${resp}")"
  fi

  actual="$(data_field "name" "${resp}")"
  if [[ "${actual}" != "${name}" ]]; then
    fail "Expected name ${name}, actual ${actual}"
  else
    pass "name matches"
  fi

  actual="$(data_field "description" "${resp}")"
  if [[ "${actual}" != "${desc}" ]]; then
    fail "Expected description '${token}', actual '${actual}'"
  else
    pass "description matches"
  fi

  actual="$(data_field "home" "${resp}")"
  if [[ "${actual}" != "${home}" ]]; then
    fail "Expected home ${home}, actual ${actual}" 
  else
    pass "home matches"
  fi

  actual="$(data_field "location" "${resp}")"
  if [[ "${actual}" != "${location}" ]]; then
    fail "Expected location ${location}, actual ${actual}"
  else
    pass "location matches"
  fi

  # Remove the player
  resp="$(players_remove "${id}")"
  if is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  report
}
