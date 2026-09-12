-- The "Overview Dashboard" feature is now simply "Dashboard". Rename the table to
-- match. SQLite rewrites the widgets.dashboard_id foreign-key reference automatically.
ALTER TABLE overview_dashboards RENAME TO dashboards;
