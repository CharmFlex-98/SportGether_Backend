-- Verify sportgether:create_event_table on pg

BEGIN;

SELECT id,
       host_id,
       event_name,
       start_time,
       end_time,
       destination,
       event_type,
       max_participant_count,
       description,
       deleted,
       version
FROM sportgether_schema.events
WHERE false;

ROLLBACK;
