package sqlite

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigration006BackfillsALegalLayout(t *testing.T) {
	const legacyWidgets = 30

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	schema, err := migrationsFS.ReadFile("migrations/004_overview_dashboards.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO overview_dashboards (id, name, "order", created_at, updated_at)
		VALUES ('d1', 'Legacy', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	for i := range legacyWidgets {
		_, err = db.Exec(`INSERT INTO widgets (id, dashboard_id, type, title, "order", config_json, created_at, updated_at)
			VALUES (?, 'd1', 'automation_list', 'Legacy', ?, '{"entity_ids":["automation.x"]}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			fmt.Sprintf("w-%02d", i), i)
		require.NoError(t, err)
	}

	migration, err := migrationsFS.ReadFile("migrations/006_overview_widget_grid.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)

	var cols, rows int
	require.NoError(t, db.QueryRow(`SELECT cols, rows FROM overview_dashboards WHERE id = 'd1'`).Scan(&cols, &rows))

	type rect struct{ col, row, colSpan, rowSpan int }
	var placed []rect

	result, err := db.Query(`SELECT col, row, col_span, row_span, json_extract(config_json, '$.display') FROM widgets`)
	require.NoError(t, err)
	defer func() { _ = result.Close() }()

	for result.Next() {
		var r rect
		var display string
		require.NoError(t, result.Scan(&r.col, &r.row, &r.colSpan, &r.rowSpan, &display))
		assert.Equal(t, "list", display)
		assert.LessOrEqual(t, r.col+r.colSpan, cols)
		assert.LessOrEqual(t, r.row+r.rowSpan, rows)
		placed = append(placed, r)
	}
	require.NoError(t, result.Err())
	require.Len(t, placed, legacyWidgets)

	for i := range placed {
		for j := i + 1; j < len(placed); j++ {
			a, b := placed[i], placed[j]
			overlaps := a.col < b.col+b.colSpan && b.col < a.col+a.colSpan &&
				a.row < b.row+b.rowSpan && b.row < a.row+a.rowSpan
			assert.Falsef(t, overlaps, "backfilled widgets %+v and %+v overlap", a, b)
		}
	}
}
