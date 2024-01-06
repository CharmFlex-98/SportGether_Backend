-- Verify sportgether:add_schema on pg

BEGIN;

SELECT pg_catalog.has_schema_privilege('sportgether_schema', 'usage');

ROLLBACK;
