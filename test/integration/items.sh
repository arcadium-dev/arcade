test_items() {
  title "Items: create, get, update and delete"

  local id
  local name="$(tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)))"
  local desc="$(for i in {1..10}; do tr -cd 'A-Za-z' < /dev/urandom 2>/dev/null | head -c $(( $RANDOM % 7  + 2)); echo -n " "; done)"
  local ownerID="00000000-0000-0000-0000-000000000001"
  local locationID="00000000-0000-0000-0000-000000000001"
  local inventoryID="00000000-0000-0000-0000-000000000001"
  local resp actual

  # Create a item
  resp="$(items_create "${name}" "${desc}" "${ownerID}" "${locationID}" "${inventoryID}")"
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

  # Check that the returned name, description, ownerID, locationID and inventoryID
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

  actual="$(data_field "locationID" "${resp}")"
  if [[ "${actual}" != "${locationID}" ]]; then
    fail "Expected locationID ${locationID}, actual ${actual}"
  else
    pass "locationID matches"
  fi

  actual="$(data_field "inventoryID" "${resp}")"
  if [[ "${actual}" != "${inventoryID}" ]]; then
    fail "Expected inventoryID ${inventoryID}, actual ${actual}"
  else
    pass "inventoryID matches"
  fi

  # Update the item name
  name="$(tr -cd 'A-Za-z' < /dev/urandom | head -c $(( $RANDOM % 7  + 2)))"
  resp="$(items_update "${id}" "${name}" "${desc}" "${ownerID}" "${locationID}" "${inventoryID}")"
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

  actual="$(data_field "locationID" "${resp}")"
  if [[ "${actual}" != "${locationID}" ]]; then
    fail "Expected locationID ${locationID}, actual ${actual}"
  else
    pass "locationID matches"
  fi

  actual="$(data_field "inventoryID" "${resp}")"
  if [[ "${actual}" != "${inventoryID}" ]]; then
    fail "Expected inventoryID ${inventoryID}, actual ${actual}"
  else
    pass "inventoryID matches"
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
