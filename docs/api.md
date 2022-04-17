```
List:   GET     /players              Get all players, filter and pagination via query params.
Get:    GET     /players/{playerID}   Get a single player.
Create: POST    /players              Create a player, w/body.
Update: UPDATE  /players/{playerID}   Update a player, w/body.
Remove: DELETE  /players/{playerID}   Delete a player.
```

The default owner for deleted items, rooms, and links is the player Nobody. (id 00000000-0000-0000-0000-000000000001).

```
List:   GET     /rooms                Get all rooms, filter and pagination via query params.
Get:    GET     /rooms/{roomID}       Get a single room.
Create: POST    /rooms                Create a room, w/body.
Update: UPDATE  /rooms/{roomID}       Update a room, w/body.
Remove: DELETE  /rooms/{roomID}       Delete a room.
```

The default room for a deleted player home, a deleted player location, and a deleted room's parent is the room Limbo (id 00000000-0000-0000-0000-000000000001).

```
List:   GET     /items                Get all items, filter and pagination via query params.
Get:    GET     /items/{itemID}       Get a single item.
Create: POST    /items                Create an item, w/body.
Update: PUT     /items/{itemID}       Update an item, w/body.
Remove: DELETE  /items/{itemsID}      Delete an item.
```

An item can be located in either a room or a player's inventory. 
If located in a room and the room is deleted, the item defaults to Limbo.
If located in a player's inventory and the player is deleted, the item defaults to Nobody.

```
List:   GET     /links                Get all links, filter and pagination via query params.
Get:    GET     /links/{linkID}       Get a single link.
Create: POST    /links                Create a link, w/body.
Update: PUT     /links/{linkID}       Update a link, w/body.
Remove: DELETE  /links/{linkID}       Delete a player.
```
