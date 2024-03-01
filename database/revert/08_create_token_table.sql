-- Revert sportgether:08_create_token_table from pg

BEGIN;

DROP TABLE sportgether_schema.tokens;

COMMIT;
