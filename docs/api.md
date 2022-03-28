```
List:   GET     /players              Get all players, filter and pagination via query params.
Get:    GET     /players/{playerID}   Get a single player.
Create: POST    /players              Create a player, w/body.
Update: UPDATE  /players/{playerID}   Update a player, w/body.
Remove: DELETE  /players/{playerID}   Delete a player. Cascade delete for all things owned by the player?
```

The NULL player is Nobody.

```
List:   GET     /rooms                Get all rooms, filter and pagination via query params.
Get:    GET     /rooms/{roomID}       Get a single room.
Create: POST    /rooms                Create a room, w/body.
Update: UPDATE  /rooms/{roomID}       Update a room, w/body.
Remove: DELETE  /rooms/{roomID}       Delete a room. All "held" things go to the owner's lost and found room?
```

The NULL room is Limbo.

```
List:   GET     /links                Get all links, filter and pagination via query params.
Get:    GET     /links/{linkID}       Get a single link.
Create: POST    /links                Create a link, w/body.
Update: PUT     /links/{linkID}       Update a link, w/body.
Remove: DELETE  /links/{linkID}       Delete a player.
```

```
List:   GET     /items                Get all items, filter and pagination via query params.
Get:    GET     /items/{itemID}       Get a single item.
Create: POST    /items                Create an item, w/body.
Update: PUT     /items/{itemID}       Update an item, w/body.
Remove: DELETE  /items/{itemsID}      Delete an item.
```
