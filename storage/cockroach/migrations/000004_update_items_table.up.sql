BEGIN;

ALTER TABLE items ADD COLUMN owner_id     UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES players (player_id) ON DELETE SET DEFAULT;
ALTER TABLE items ADD COLUMN location_id  UUID          DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES rooms   (room_id)   ON DELETE SET DEFAULT;
ALTER TABLE items ADD COLUMN inventory_id UUID                                                         REFERENCES players (player_id) ON DELETE SET NULL;

CREATE INDEX items_by_owner_index     ON items (owner_id);
CREATE INDEX items_in_location_index  ON items (location_id);
CREATE INDEX items_in_inventory_index ON items (inventory_id);

COMMIT;
