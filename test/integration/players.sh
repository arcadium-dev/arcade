test_players() {
  title "Players: create, get, update and delete"

  local name="$(tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)))"
  local desc="$(for i in {1..10}; do tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)); echo -n " "; done)"
  local homeID="00000000-0000-0000-0000-000000000001"
  local locationID="00000000-0000-0000-0000-000000000001"
  local resp id actual

  # Create a player
  resp="$(players_create "${name}" "${desc}" "${homeID}" "${locationID}")"
  if [[ $? -ne $SUCCESS ]]; then
    fatal "Fail: ${resp}"
  elif is_error "${resp}"; then
    fatal "$(error_detail "${resp}")"
  fi
  id="$(data_field "playerID" "${resp}")"


  # Get the player
  resp="$(players_get "${id}")"
  if [[ $? -ne $SUCCESS ]]; then
    fail "Fail: ${resp}"
  elif is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  # Check that the returned name, description, homeID and locationID
  actual="$(data_field "name" "${resp}")"
  if [[ "${actual}" != "${name}" ]]; then
    fail "Expected name ${name}, actual ${actual}"
  else
    pass "name matches"
  fi

  actual="$(data_field "description" "${resp}")"
  if [[ "${actual}" != "${desc}" ]]; then
    fail "Expected description '${desc}', actual '${actual}'"
  else
    pass "description matches"
  fi

  actual="$(data_field "homeID" "${resp}")"
  if [[ "${actual}" != "${homeID}" ]]; then
    fail "Expected homeID ${homeID}, actual ${actual}"
  else
    pass "homeID matches"
  fi

  actual="$(data_field "locationID" "${resp}")"
  if [[ "${actual}" != "${locationID}" ]]; then
    fail "Expected locationID ${locationID}, actual ${actual}"
  else
    pass "locationID matches"
  fi

  # Update the player name
  name="$(tr -cd 'A-Za-z' < /dev/urandom | head -c $(( $RANDOM % 7  + 2)))"
  resp="$(players_update "${id}" "${name}" "${desc}" "${homeID}" "${locationID}")"
  if [[ $? -ne $SUCCESS ]]; then
    fatal "Fail: ${resp}"
  elif is_error "${resp}"; then
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
    fail "Expected description '${desc}', actual '${actual}'"
  else
    pass "description matches"
  fi

  actual="$(data_field "homeID" "${resp}")"
  if [[ "${actual}" != "${homeID}" ]]; then
    fail "Expected homeID ${homeID}, actual ${actual}" 
  else
    pass "homeID matches"
  fi

  actual="$(data_field "locationID" "${resp}")"
  if [[ "${actual}" != "${locationID}" ]]; then
    fail "Expected locationID ${locationID}, actual ${actual}"
  else
    pass "locationID matches"
  fi

  # Remove the player
  resp="$(players_remove "${id}")"
  if [[ $? -ne $SUCCESS ]]; then
    fatal "Fail: ${resp}"
  elif is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  report
}
