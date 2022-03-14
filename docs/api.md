```
List:   GET     /players              Get all players, filter and pagination via query params
Get:    GET     /players/{playerID}   Get a single player
Create: POST    /players              Create a player, w/body
Update: UPDATE  /players/{playerID}   Update a player, w/body
Remove: DELETE  /players/{playerID}   Delete a player. Cascade delete for all things owned by the player?
```

```
GET     /rooms                Get all rooms, filter and pagination via query params
GET     /rooms/{roomID}       Get a single room
POST    /rooms                Create a room, w/body
UPDATE  /rooms/{roomID}       Update a room, w/body
DELETE  /rooms/{roomID}       Delete a room. All "held" things go to the owner's lost and found room?
```

GET     /links
GET     /links/{linkID}
POST    /links
PUT     /links/{linkID}
DELETE  /links/{linkID}

GET     /items
GET     /items/{itemID}
POST    /items
PUT     /items/{itemID}
DELETE  /items/{itemsID}
