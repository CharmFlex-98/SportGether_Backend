-- Deploy sportgether:create_event_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.events
(
    id                    bigserial PRIMARY KEY,
    event_name            text NOT NULL,
    start_time            text NOT NULL,
    end_time              text NOT NULL,
    destination           text NOT NULL,
    event_type            text NOT NULL,
    max_participant_count int  NOT NULL,
    description           text
);

COMMIT;
