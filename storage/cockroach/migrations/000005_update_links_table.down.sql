BEGIN;

DROP INDEX IF EXISTS links_to_destination_index;
DROP INDEX IF EXISTS links_from_location_index;
DROP INDEX IF EXISTS links_by_owner_index;

ALTER TABLE links DROP COLUMN destination_id;
ALTER TABLE links DROP COLUMN location_id;
ALTER TABLE links DROP COLUMN owner_id;

COMMIT;
