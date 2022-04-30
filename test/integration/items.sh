test_items() {
  title "Items: create, get, update and delete"

  local id
  local name="$(tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)))"
  local desc="$(for i in {1..10}; do tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)); echo -n " "; done)"
  local owner="00000000-0000-0000-0000-000000000001"
  local location="00000000-0000-0000-0000-000000000001"
  local inventory="00000000-0000-0000-0000-000000000001"
  local resp actual

  # Create a item
  resp="$(items_create "${name}" "${desc}" "${owner}" "${location}" "${inventory}")"
  if [[ $? -ne $SUCCESS ]]; then
    fatal "Fail: ${resp}"
  elif is_error "${resp}"; then
    fatal "$(error_detail "${resp}")"
  fi
  id="$(data_field "itemID" "${resp}")"


  # Get the item
  resp="$(items_get "${id}")"
  if [[ $? -ne $SUCCESS ]]; then
    fail "Fail: ${resp}"
  elif is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  # Check that the returned name, description, owner, location and inventory
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

  actual="$(data_field "owner" "${resp}")"
  if [[ "${actual}" != "${owner}" ]]; then
    fail "Expected owner ${owner}, actual ${actual}"
  else
    pass "owner matches"
  fi

  actual="$(data_field "location" "${resp}")"
  if [[ "${actual}" != "${location}" ]]; then
    fail "Expected location ${location}, actual ${actual}"
  else
    pass "location matches"
  fi

  actual="$(data_field "inventory" "${resp}")"
  if [[ "${actual}" != "${inventory}" ]]; then
    fail "Expected inventory ${inventory}, actual ${actual}"
  else
    pass "inventory matches"
  fi

  # Update the item name
  name="$(tr -cd 'A-Za-z' < /dev/urandom | head -c $(( $RANDOM % 7  + 2)))"
  resp="$(items_update "${id}" "${name}" "${desc}" "${owner}" "${location}" "${inventory}")"
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

  actual="$(data_field "owner" "${resp}")"
  if [[ "${actual}" != "${owner}" ]]; then
    fail "Expected owner ${owner}, actual ${actual}" 
  else
    pass "owner matches"
  fi

  actual="$(data_field "location" "${resp}")"
  if [[ "${actual}" != "${location}" ]]; then
    fail "Expected location ${location}, actual ${actual}"
  else
    pass "location matches"
  fi

  actual="$(data_field "inventory" "${resp}")"
  if [[ "${actual}" != "${inventory}" ]]; then
    fail "Expected inventory ${inventory}, actual ${actual}"
  else
    pass "inventory matches"
  fi

  # Remove the item
  resp="$(items_remove "${id}")"
  if [[ $? -ne $SUCCESS ]]; then
    fatal "Fail: ${resp}"
  elif is_error "${resp}"; then
    fail "$(error_detail "${resp}")"
  fi

  report
}
