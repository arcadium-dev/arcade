BEGIN;

CREATE TABLE IF NOT EXISTS items (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE IF NOT EXISTS links (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE IF NOT EXISTS players (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT UNIQUE NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE IF NOT EXISTS rooms (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT NOT NULL,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

COMMIT;

BEGIN;

INSERT INTO items (id, name, description)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Nothing',
  'A thing of no importance.'
);

INSERT INTO players (id, name, description)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Nobody',
  'A person of no importance.'
);

INSERT INTO rooms (id, name, description) 
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Nowhere',
  'A place of no importance.'
);

COMMIT;
