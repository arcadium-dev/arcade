BEGIN;

DROP INDEX IF EXISTS players_in_location_index;

ALTER TABLE players DROP COLUMN location;
ALTER TABLE players DROP COLUMN home;

COMMIT;
