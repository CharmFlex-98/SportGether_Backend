-- Verify sportgether:create_user_profile_table on pg

BEGIN;

SELECT user_id,
       preferred_name,
       gender,
       birthday,
       join_date,
       profile_icon_url,
       signature,
       memo
FROM sportgether_schema.user_profile
WHERE false;

ROLLBACK;
