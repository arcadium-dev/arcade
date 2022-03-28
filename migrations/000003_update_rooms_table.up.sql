BEGIN;

ALTER TABLE rooms ADD COLUMN owner  UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES players (player_id) ON DELETE SET DEFAULT;
ALTER TABLE rooms ADD COLUMN parent UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES rooms   (room_id)   ON DELETE SET DEFAULT;

CREATE INDEX rooms_by_owner_index  ON rooms (owner);
CREATE INDEX rooms_in_parent_index ON rooms (parent);

COMMIT;

UPDATE rooms SET
  owner  = '00000000-0000-0000-0000-000000000001',
  parent = '00000000-0000-0000-0000-000000000001'
WHERE
  room_id = '00000000-0000-0000-0000-000000000001';
