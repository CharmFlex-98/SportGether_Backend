-- Verify sportgether:08_create_token_table on pg

BEGIN;

SELECT hash, 
    user_id, 
    expiry,
    scope text
FROM sportgether_schema.tokens
WHERE false;
ROLLBACK;
