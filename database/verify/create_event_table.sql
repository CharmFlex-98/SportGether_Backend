-- Verify sportgether:create_event_table on pg

BEGIN;

SELECT id,
       event_name,
       start_time,
       end_time,
       destination,
       event_type,
       max_participant_count,
       description
FROM sportgether_schema.events
WHERE false;

ROLLBACK;
