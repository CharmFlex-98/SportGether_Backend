-- Deploy sportgether:create_event_participant_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.event_participant(
    eventId bigint REFERENCES sportgether_schema.events ON DELETE CASCADE,
    participantId bigint REFERENCES sportgether_schema.users ON DELETE CASCADE
);

COMMIT;
