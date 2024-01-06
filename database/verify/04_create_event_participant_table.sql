-- Verify sportgether:create_event_participant_table on pg

BEGIN;

SELECT eventId, participantId from sportgether_schema.event_participant WHERE FALSE;
ROLLBACK;
