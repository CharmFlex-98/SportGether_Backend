-- Deploy sportgether:create_event_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.events
(
    id                    bigserial PRIMARY KEY,
    host_id bigint NOT NULL REFERENCES sportgether_schema.users on DELETE CASCADE,
    event_name            text NOT NULL,
    start_time            timestamp(0) with time zone NOT NULL,
    end_time              timestamp(0) with time zone NOT NULL,
    destination           text NOT NULL,
    event_type            text NOT NULL,
    max_participant_count int  NOT NULL,
    description           text
);

COMMIT;
