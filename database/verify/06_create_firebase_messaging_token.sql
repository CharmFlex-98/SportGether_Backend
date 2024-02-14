-- Verify sportgether:06_create_firebase_messaging_token on pg

BEGIN;

SELECT user_id,
       token
FROM sportgether_schema.firebase_messaging_token_table
WHERE false;

ROLLBACK;
