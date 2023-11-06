-- Revert sportgether:add_schema from pg

BEGIN;

DROP SCHEMA sportgether_schema;

COMMIT;
