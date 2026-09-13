package dashboard

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// exportedSnapshot fetches the current configuration through GET /api/export as a generic map, so
// a test can tamper with individual fields before replaying it through POST /api/restore.
func exportedSnapshot(t *testing.T, owner *authTestClient) map[string]any {
	t.Helper()
	resp := owner.do(http.MethodGet, "/api/export", nil, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, "export: %s", body)

	// The endpoint answers with a JSend envelope; the snapshot itself is the data member.
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &envelope))
	require.NotNil(t, envelope.Data)
	return envelope.Data
}

// findDashboard returns the exported dashboard with the given name.
func findDashboard(t *testing.T, snapshot map[string]any, name string) map[string]any {
	t.Helper()
	for _, raw := range snapshot["dashboards"].([]any) {
		dashboard := raw.(map[string]any)
		if dashboard["name"] == name {
			return dashboard
		}
	}
	t.Fatalf("dashboard %q not found in the snapshot", name)
	return nil
}

// restoreSnapshot replays a snapshot and returns the status code and body.
func restoreSnapshot(t *testing.T, owner *authTestClient, snapshot map[string]any) (int, string) {
	t.Helper()
	payload := map[string]any{
		"version":         snapshot["version"],
		"levels":          snapshot["levels"],
		"dashboards":      snapshot["dashboards"],
		"include_devices": false,
	}
	resp := owner.do(http.MethodPost, "/api/restore", payload, nil)
	body := readBody(t, resp)
	return resp.StatusCode, body
}

// seedExportableConfiguration builds one level with a plan, and one dashboard in a 12x8 grid
// holding a 2x2 widget, so the exported snapshot exercises every branch of the restore path.
func seedExportableConfiguration(t *testing.T, owner *authTestClient) (levelID string) {
	t.Helper()
	levelID = createLevel(t, owner, "Ground floor")

	resp := owner.do(http.MethodPost, "/api/levels/"+levelID+"/plan/import", map[string]any{
		"walls": []map[string]any{
			{"start": map[string]float64{"x": 0, "y": 0}, "end": map[string]float64{"x": 5.2, "y": 0}, "thickness_cm": 20},
			{"start": map[string]float64{"x": 5.2, "y": 0}, "end": map[string]float64{"x": 5.2, "y": 4.3}, "thickness_cm": 20},
		},
		"zones": []any{},
	}, nil)
	planBody := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, "import plan: %s", planBody)

	resp = owner.do(http.MethodPost, "/api/dashboards", map[string]any{"name": "Overview"}, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, "create dashboard: %s", body)
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &created))

	resp = owner.do(http.MethodPost, "/api/dashboards/"+created.Data.ID+"/widgets", map[string]any{
		"type":     "automation_list",
		"title":    "Scenes",
		"col":      10,
		"row":      0,
		"col_span": 2,
		"row_span": 2,
		"config":   map[string]any{"entity_ids": []string{"automation.cinema"}, "display": "list"},
	}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "create widget: %s", readBody(t, resp))

	return levelID
}

// TestRestoreRejectsDanglingParents is the regression test for the audit finding
// "POST /api/restore persists client-chosen primary keys, timestamps and parent ids". A restore
// replays a snapshot as a graph, so references inside it must resolve against the snapshot:
// a plan or placement pointing at a level the snapshot does not contain used to be written
// verbatim and then served by GET /api/levels/{id}/plan for a level that does not exist.
func TestRestoreRejectsDanglingParents(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	levelID := seedExportableConfiguration(t, owner)

	snapshot := exportedSnapshot(t, owner)
	levels := snapshot["levels"].([]any)
	require.Len(t, levels, 1)
	plan := levels[0].(map[string]any)["plan"].(map[string]any)
	plan["level_id"] = "ghost-level-id"

	code, body := restoreSnapshot(t, owner, snapshot)
	require.Equal(t, http.StatusBadRequest, code,
		"a snapshot whose plan references an unknown level must be rejected, got %d: %s", code, body)
	require.Contains(t, body, "INVALID_PLAN")

	// Nothing was replayed: the original level and its plan are intact.
	resp := owner.do(http.MethodGet, "/api/levels/"+levelID+"/plan", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "the previous configuration must survive a rejected restore")
	resp.Body.Close()
}

// TestRestoreRejectsWidgetNestedUnderAnotherDashboard covers the second half of the same finding:
// a widget listed inside one dashboard while naming another was stored on the named dashboard
// without ever being validated against that dashboard's grid, so a 2x2 widget at column 10 ended
// up on a 4x4 grid — a state the ordinary widget route refuses with INVALID_WIDGET.
func TestRestoreRejectsWidgetNestedUnderAnotherDashboard(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	levelID := seedExportableConfiguration(t, owner)
	_ = levelID

	// A second, deliberately small grid that the widget cannot fit into.
	resp := owner.do(http.MethodPost, "/api/dashboards", map[string]any{
		"name": "Small", "cols": 4, "rows": 4,
	}, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, "create small dashboard: %s", body)
	var small struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &small))

	snapshot := exportedSnapshot(t, owner)

	// Keep the widget nested where it is (valid for the 12x8 grid it is listed under) but point
	// it at the 4x4 grid, where a 2x2 widget at column 10 cannot fit.
	overview := findDashboard(t, snapshot, "Overview")
	widgets := overview["widgets"].([]any)
	require.NotEmpty(t, widgets)
	widgets[0].(map[string]any)["dashboard_id"] = small.Data.ID

	code, body := restoreSnapshot(t, owner, snapshot)
	require.Equal(t, http.StatusBadRequest, code,
		"a widget must be validated against the dashboard it names, got %d: %s", code, body)
	require.Contains(t, body, "INVALID_WIDGET")
}

// TestRestoreAcceptsItsOwnExport is the control: a snapshot produced by GET /api/export must replay
// unchanged, so the new graph validation does not reject legitimate restores.
func TestRestoreAcceptsItsOwnExport(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	levelID := seedExportableConfiguration(t, owner)
	snapshot := exportedSnapshot(t, owner)

	code, body := restoreSnapshot(t, owner, snapshot)
	require.Equal(t, http.StatusOK, code, "an unmodified export must restore: %s", body)
	require.Contains(t, body, `"status":"success"`)

	resp := owner.do(http.MethodGet, "/api/levels/"+levelID+"/plan", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	names := make([]string, 0)
	for _, dashboard := range listDashboards(t, owner) {
		names = append(names, dashboard["name"].(string))
	}
	require.Contains(t, names, "Overview", "the restored dashboard must be present")
}

// TestRestoreRejectsOversizedLevelName covers the last part of the finding: the restore path
// skipped the DTO limits the ordinary routes enforce, so a 500-character level name that
// POST /api/levels rejects with 400 was accepted verbatim by a restore.
func TestRestoreRejectsOversizedLevelName(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	seedExportableConfiguration(t, owner)
	snapshot := exportedSnapshot(t, owner)

	levels := snapshot["levels"].([]any)
	level := levels[0].(map[string]any)["level"].(map[string]any)
	level["name"] = string(make([]rune, 500))

	code, body := restoreSnapshot(t, owner, snapshot)
	require.Equal(t, http.StatusBadRequest, code,
		"a restore must apply the same name limits as the ordinary routes, got %d: %s", code, body)
}

// listDashboards returns the dashboards as raw maps.
func listDashboards(t *testing.T, owner *authTestClient) []map[string]any {
	t.Helper()
	resp := owner.do(http.MethodGet, "/api/dashboards", nil, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &out))
	return out.Data
}
