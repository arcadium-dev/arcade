BEGIN;

ALTER TABLE links ADD COLUMN owner_id       UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES players (player_id) ON DELETE SET DEFAULT;
ALTER TABLE links ADD COLUMN location_id    UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES rooms   (room_id)   ON DELETE SET DEFAULT;
ALTER TABLE links ADD COLUMN destination_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES rooms   (room_id)   ON DELETE SET DEFAULT;

CREATE INDEX links_by_owner_index       ON links (owner_id);
CREATE INDEX links_from_location_index  ON links (location_id);
CREATE INDEX links_to_destination_index ON links (destination_id);

COMMIT;
