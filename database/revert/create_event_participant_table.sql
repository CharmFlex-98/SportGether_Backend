-- Revert sportgether:create_event_participant_table from pg

BEGIN;

DROP TABLE sportgether_schema.event_participant;
COMMIT;
