BEGIN;

DROP INDEX IF EXISTS links_to_destination_index;
DROP INDEX IF EXISTS links_from_location_index;
DROP INDEX IF EXISTS links_by_owner_index;

ALTER TABLE links DROP COLUMN destination;
ALTER TABLE links DROP COLUMN location;
ALTER TABLE links DROP COLUMN owner;

COMMIT;
