ALTER TABLE overview_dashboards ADD COLUMN cols INTEGER NOT NULL DEFAULT 12;
ALTER TABLE overview_dashboards ADD COLUMN rows INTEGER NOT NULL DEFAULT 8;
ALTER TABLE widgets ADD COLUMN col INTEGER NOT NULL DEFAULT 0;
ALTER TABLE widgets ADD COLUMN row INTEGER NOT NULL DEFAULT 0;
ALTER TABLE widgets ADD COLUMN col_span INTEGER NOT NULL DEFAULT 2;
ALTER TABLE widgets ADD COLUMN row_span INTEGER NOT NULL DEFAULT 2;

-- Pack legacy widgets row-major as 2x2 tiles, six per grid row, so that no two of them
-- overlap. Overlapping widgets are rejected by OverviewDashboard.Validate, which would
-- otherwise make every later write on a migrated dashboard fail.
UPDATE widgets
SET col = ranked.new_col,
    row = ranked.new_row
FROM (
    SELECT id,
           2 * ((rn - 1) % 6) AS new_col,
           2 * ((rn - 1) / 6) AS new_row
    FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY dashboard_id ORDER BY "order", created_at, id) AS rn
        FROM widgets
    )
) AS ranked
WHERE widgets.id = ranked.id;

-- Six 2x2 widgets per grid row means the default eight rows only hold twenty-four of them.
-- Grow the grid of any dashboard that carried more, rather than leave a widget outside the
-- viewport: cells stretch, the dashboard still never scrolls.
UPDATE overview_dashboards
SET rows = MAX(8, packed.needed_rows)
FROM (
    SELECT dashboard_id, 2 * ((COUNT(*) + 5) / 6) AS needed_rows
    FROM widgets
    GROUP BY dashboard_id
) AS packed
WHERE overview_dashboards.id = packed.dashboard_id;

-- Every widget written before this migration was an automation list: that was the only kind
-- the UI could create. Map it onto the one Display Mode its type accepts, and give any other
-- legacy type the display its own type accepts, so no row fails Widget.Validate.
UPDATE widgets
SET config_json = json_set(config_json, '$.display',
    CASE type
        WHEN 'automation_list' THEN 'list'
        WHEN 'actuator' THEN 'toggle'
        WHEN 'sensor' THEN 'number'
        ELSE 'list'
    END)
WHERE json_extract(config_json, '$.display') IS NULL;
