-- Revert sportgether:remove_app_schema from pg

BEGIN;

CREATE SCHEMA sportgether;

COMMIT;
