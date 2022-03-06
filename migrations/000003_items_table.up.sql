CREATE TABLE IF NOT EXISTS items (
  items_id UUID PRIMARY KEY,

  created TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
  updated TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);
