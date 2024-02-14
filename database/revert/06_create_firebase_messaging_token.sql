-- Revert sportgether:06_create_firebase_messaging_token from pg

BEGIN;

DROP TABLE sportgether_schema.firebase_messaging_token_table;

COMMIT;
