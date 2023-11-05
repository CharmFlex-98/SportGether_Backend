-- Verify sportgether:appschema on pg

BEGIN;

SELECT pg_catalog.has_schema_privilege('sportgether', 'usage');

ROLLBACK;
