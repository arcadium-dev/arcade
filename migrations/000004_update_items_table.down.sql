BEGIN;

DROP INDEX IF EXISTS items_in_inventory_index;
DROP INDEX IF EXISTS items_in_location_index;
DROP INDEX IF EXISTS items_by_owner_index;

ALTER TABLE items DROP COLUMN inventory;
ALTER TABLE items DROP COLUMN location;
ALTER TABLE items DROP COLUMN owner;

COMMIT;
