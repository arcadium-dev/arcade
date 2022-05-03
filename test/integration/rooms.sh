test_rooms() {
  title "Rooms: create, get, update and delete"

  local id
  local name="$(tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)))"
  local desc="$(for i in {1..10}; do tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)); echo -n " "; done)"
  local ownerID="00000000-0000-0000-0000-000000000001"
  local parentID="00000000-0000-0000-0000-000000000001"
  local resp actual

  # Create a room
  resp="$(rooms_create "${name}" "${desc}" "${ownerID}" "${parentID}")"
  if [[ $? -ne $SUCCESS ]]; then
    fatal "Fail: ${resp}"
  elif is_error "${resp}"; then
    fatal "$(error_detail "${resp}")"
  fi
  id="$(data_field "roomID" "${resp}")"


  # Get the room
  resp="$(rooms_get "${id}")"
  if [[ $? -ne $SUCCESS ]]; then
    fail "Fail: ${resp}"
  elif is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  # Check that the returned name, description, ownerID and parentID
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

  actual="$(data_field "ownerID" "${resp}")"
  if [[ "${actual}" != "${ownerID}" ]]; then
    fail "Expected ownerID ${ownerID}, actual ${actual}"
  else
    pass "ownerID matches"
  fi

  actual="$(data_field "parentID" "${resp}")"
  if [[ "${actual}" != "${parentID}" ]]; then
    fail "Expected parentID ${parentID}, actual ${actual}"
  else
    pass "parentID matches"
  fi

  # Update the room name
  name="$(tr -cd 'A-Za-z' < /dev/urandom | head -c $(( $RANDOM % 7  + 2)))"
  resp="$(rooms_update "${id}" "${name}" "${desc}" "${ownerID}" "${parentID}")"
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

  actual="$(data_field "ownerID" "${resp}")"
  if [[ "${actual}" != "${ownerID}" ]]; then
    fail "Expected ownerID ${ownerID}, actual ${actual}"
  else
    pass "ownerID matches"
  fi

  actual="$(data_field "parentID" "${resp}")"
  if [[ "${actual}" != "${parentID}" ]]; then
    fail "Expected parentID ${parentID}, actual ${actual}"
  else
    pass "parentID matches"
  fi

  # Remove the room
  resp="$(rooms_remove "${id}")"
  if [[ $? -ne $SUCCESS ]]; then
    fatal "Fail: ${resp}"
  elif is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  report
}
