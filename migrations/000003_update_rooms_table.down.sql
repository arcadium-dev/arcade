BEGIN;

DROP INDEX IF EXISTS rooms_in_parent_index;
DROP INDEX IF EXISTS rooms_by_owner_index;

ALTER TABLE rooms DROP COLUMN parent;
ALTER TABLE rooms DROP COLUMN owner;

COMMIT;
