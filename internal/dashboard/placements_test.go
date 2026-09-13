package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// createLevel creates a level through the API and returns its id.
func createLevel(t *testing.T, client *authTestClient, name string) string {
	t.Helper()
	resp := client.do(http.MethodPost, "/api/levels", map[string]any{"name": name}, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, "create level %q: %s", name, body)

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &created))
	require.NotEmpty(t, created.Data.ID)
	return created.Data.ID
}

// placeDevice saves a placement through the API and returns its id.
func placeDevice(t *testing.T, client *authTestClient, levelID, deviceID, customName string) string {
	t.Helper()
	resp := client.do(http.MethodPost, "/api/levels/"+levelID+"/placements", map[string]any{
		"device_id":   deviceID,
		"x":           10,
		"y":           20,
		"custom_name": customName,
	}, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, "place device: %s", body)

	var placed struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &placed))
	require.NotEmpty(t, placed.Data.ID)
	return placed.Data.ID
}

// listPlacements returns the placements of a level as raw maps, so a test can assert on the
// persisted values rather than on a typed subset.
func listPlacements(t *testing.T, client *authTestClient, levelID string) []map[string]any {
	t.Helper()
	resp := client.do(http.MethodGet, "/api/levels/"+levelID+"/placements", nil, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &out))
	return out.Data
}

// TestPlacementCannotBeMovedAcrossLevels is the regression test for the audit finding
// "Client-supplied placement id on POST /api/levels/{id}/placements rewrites another level's
// placement". The request DTO carries an optional id that becomes the primary key of an upsert,
// so naming a placement that belongs to a different level used to move that row onto the target
// level and overwrite its contents. The sibling DELETE route already refuses the same mismatch,
// and the write route must agree with it.
func TestPlacementCannotBeMovedAcrossLevels(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	levelA := createLevel(t, owner, "Ground floor")
	levelB := createLevel(t, owner, "First floor")

	placementID := placeDevice(t, owner, levelB, "light.kitchen", "Kitchen counter")

	// Name level B's placement while posting to level A: the row must not move.
	resp := owner.do(http.MethodPost, "/api/levels/"+levelA+"/placements", map[string]any{
		"id":          placementID,
		"device_id":   "switch.other",
		"x":           1,
		"y":           2,
		"custom_name": "STOLEN",
	}, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"a placement must not be moved to another level, got %d: %s", resp.StatusCode, body)
	require.Contains(t, body, "PLACEMENT_NOT_FOUND")

	// Level B still owns its placement, with its original contents.
	placementsB := listPlacements(t, owner, levelB)
	require.Len(t, placementsB, 1, "level B must keep its placement")
	require.Equal(t, placementID, placementsB[0]["id"])
	require.Equal(t, "light.kitchen", placementsB[0]["device_id"])
	require.Equal(t, "Kitchen counter", placementsB[0]["custom_name"])

	// ...and level A gained nothing.
	require.Empty(t, listPlacements(t, owner, levelA))

	// Updating the placement in place, through its own level, still works.
	resp = owner.do(http.MethodPost, "/api/levels/"+levelB+"/placements", map[string]any{
		"id":          placementID,
		"device_id":   "light.kitchen",
		"x":           33,
		"y":           44,
		"custom_name": "Moved on its own plan",
	}, nil)
	updateBody := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, "in-place update must keep working: %s", updateBody)

	placementsB = listPlacements(t, owner, levelB)
	require.Len(t, placementsB, 1)
	require.Equal(t, "Moved on its own plan", placementsB[0]["custom_name"])
	require.InDelta(t, 33, placementsB[0]["x"], 0.001, fmt.Sprintf("x should be 33, got %v", placementsB[0]["x"]))
}
