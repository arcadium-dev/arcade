BEGIN;

ALTER TABLE items ADD COLUMN owner_id           UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES players (id) ON DELETE SET DEFAULT;
ALTER TABLE items ADD COLUMN location_item_id   UUID                                                         REFERENCES items   (id) ON DELETE SET DEFAULT;
ALTER TABLE items ADD COLUMN location_player_id UUID                                                         REFERENCES players (id) ON DELETE SET DEFAULT;
ALTER TABLE items ADD COLUMN location_room_id   UUID                                                         REFERENCES rooms   (id) ON DELETE SET DEFAULT;

CREATE INDEX items_by_owner_index  ON items (owner_id);
CREATE INDEX items_in_item_index   ON items (location_item_id);
CREATE INDEX items_in_player_index ON items (location_player_id);
CREATE INDEX items_in_room_index   ON items (location_room_id);

COMMIT;

UPDATE items SET
  owner_id           = '00000000-0000-0000-0000-000000000001',
  location_player_id = '00000000-0000-0000-0000-000000000001'
WHERE
  id = '00000000-0000-0000-0000-000000000001';
