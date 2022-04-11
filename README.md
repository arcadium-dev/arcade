# arcade

export player_id=$(uuidgen)
export player_name=Somebody
export player_desc="A person of importance."
export home_id=$(uuidgen)
export location_id=$(uuidgen)


dev run curl --request POST --data '{
  "playerID": "'${player_id}'", "name": "'${player_name}'", "description": "'${player_desc}'", "home": "'${home_id}'", "location": "'${location_id}'"
}' "https://assets:4201/players"

dev run curl --request GET "https://assets:4201/players"

dev run curl --request GET "https://assets:4201/players/${player_id}"
