# Design

## Commands

say
  `say ...text...`
  `" ...text...`

  say hello world!
  " hello world!

  Bob says, "hello world!"

emote
  `emote ...text...`
  `: ...text...`

  emote dances as if no one is looking.
  : dances as if no one is looking.

  Bob dances as if no one is looking.

look 

examine

use


## Game Objects

### Player

#### Attributes

Name
  The name of the player. This must be unique across all players.

Description
  The description of the player.

Location
  This must be a room.
  The location of the player. 
  The default location of a player is "Limbo". This is configurable by the admin.
    This is the initial location of the player.
    If a room is deleted, the players in the room will be moved to "Limbo".

Link
  This must be a room.
  The linked room is the player's home.
  When a player logs out, they will be moved to their "Home".
  The default home of a player is "Limbo". This is configurable by the admin.
    This is the initial home of a player.
    If a player's home is deleted, the players home will default to "Limbo".

Flags
  A - admin
  B - builder


#### Properties

#### Actions

@create
  `@create name[=description]`

  Creates a new player. 
  Can only be performed by an admin.
  A default description is provided if the description is not given.

@boot
  `@boot name`

  Disconnects the player, as if they quit.
  Can only be performed by an admin.
  The player is informed they have been booted.

@ban
  `@ban name`

  Bans a player. 
  Can only be performed by an admin.
  The player is not allowed to login.
  The player object is not destroyed.

@unban
  `@unban name`

  Unbans a player.
  Can only be performed by an admin.
  The player is allowed to login.

@destroy
  `destroy name[=owner]`

  The player object is destroyed.
  If the owner is specified, all things owned by the player will tranfer ownership to the new owner.
  If the owner is not specified, all things owned by the player will transfer to "Nobody". This default can be configured by the admin.

@name
  `@name new name`

  Renames a player. 
  The new player name must be unique.

@desc
  `@desc new description`

  Updates the player description.

@link
  `@link room`

  Sets a player's home room.



### Room

#### Attributes

Name
  The name of the room. This is not required to be unique.

Description
  The description of the room.

Owner
  The player id of the owner of the room.
  This is the id of the player who dug the room.
  If the player is destroyed the owner may be explicitly set to a new player, or it may default to "Nobody". This default can be configured by the admin.

Location
  This must be a room. Rooms can be located in other rooms, i.e a hierachy of environments.
    The command search algorithm will search through the hierarchy of rooms to attempt to match a command.
  The default location of a room is "The Universe". This default room location is configurable by the admin.

Link
  A room can be linked to another room.

Flags
  D - dark. The contents of the room, the players and items located in this room, are not visible to other players.
  H - home. Any player can link to this room and set it their home.
  Q - quiet. Notifications will not be generated from this room.
  S - sticky. 


#### Properties


### Item

#### Attributes

#### Properties


### Link

#### Attributes

#### Properties



