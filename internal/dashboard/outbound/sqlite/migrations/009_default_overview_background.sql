-- Give every existing Overview Dashboard the bundled default background. Dashboards
-- that explicitly opted out (bg_image stays '') are unaffected; only rows that predate
-- the background feature (empty string) are backfilled.
UPDATE overview_dashboards
SET bg_image = '/backgrounds/desert-night.jpg'
WHERE bg_image = '';
