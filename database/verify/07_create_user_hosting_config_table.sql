-- Verify sportgether:07_create_user_hosting_config_table on pg

BEGIN;

SELECT user_id, 
    host_count, 
    last_refresh_time,
    challenge_id text
FROM sportgether_schema.user_hosting_config
WHERE false;

ROLLBACK;
