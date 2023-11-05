-- Revert sportgether:appschema from pg

BEGIN;

DROP SCHEMA sportgether;

COMMIT;
