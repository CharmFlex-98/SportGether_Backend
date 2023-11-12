-- Revert sportgether:create_token_table from pg

BEGIN;

DROP TABLE IF EXISTS sportgether_schema.tokens;

COMMIT;
