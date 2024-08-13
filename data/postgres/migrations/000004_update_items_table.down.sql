BEGIN;

DROP INDEX IF EXISTS items_in_room_index;
DROP INDEX IF EXISTS items_in_player_index;
DROP INDEX IF EXISTS items_in_item_index;
DROP INDEX IF EXISTS items_by_owner_index;

ALTER TABLE items DROP COLUMN location_room_id;
ALTER TABLE items DROP COLUMN location_player_id;
ALTER TABLE items DROP COLUMN location_item_id;
ALTER TABLE items DROP COLUMN owner_id;

COMMIT;
