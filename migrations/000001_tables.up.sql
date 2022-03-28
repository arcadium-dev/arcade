BEGIN;

CREATE TABLE IF NOT EXISTS players (
  player_id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name        TEXT NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE IF NOT EXISTS rooms (
  room_id    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name        TEXT NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE IF NOT EXISTS items (
  item_id    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name        TEXT NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE IF NOT EXISTS links (
  link_id    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name        TEXT NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

INSERT INTO players (player_id, name, description)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Nobody',
  'A person of no importance.'
);

INSERT INTO rooms (room_id, name, description) 
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Limbo',
  'A place for lost, forgotten, or unwanted persons or things.'
);

COMMIT;
