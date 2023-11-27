-- Revert sportgether:create_event_table from pg

BEGIN;

DROP TABLE sportgether_schema.events;

COMMIT;
