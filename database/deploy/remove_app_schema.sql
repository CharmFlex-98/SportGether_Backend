-- Deploy sportgether:remove_app_schema to pg

BEGIN;

DROP SCHEMA sportgether;

COMMIT;
