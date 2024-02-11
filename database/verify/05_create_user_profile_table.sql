-- Verify sportgether:create_user_profile_table on pg

BEGIN;

SELECT user_id,
       preferred_name,
       gender,
       birth_date,
       join_date,
       status,
       profile_icon_url,
       signature,
       memo,
       version
FROM sportgether_schema.user_profile
WHERE false;

ROLLBACK;
