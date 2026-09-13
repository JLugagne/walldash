package aijson_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/aijson"
	"github.com/stretchr/testify/require"
)

// planFixture builds an AI-plan JSON document with enough walls and zones to derive dozens of ids
// from a single import.
func planFixture(t testing.TB) []byte {
	t.Helper()
	walls := make([]map[string]any, 0, 60)
	for i := 0; i < 60; i++ {
		walls = append(walls, map[string]any{
			"start":        map[string]float64{"x": float64(i), "y": 0},
			"end":          map[string]float64{"x": float64(i) + 0.8, "y": 0},
			"thickness_cm": 10,
			"openings": []map[string]any{
				{"type": "door", "offset_m": float64(i) + 0.4, "width_cm": 80},
			},
		})
	}
	zones := make([]map[string]any, 0, 10)
	for i := 0; i < 10; i++ {
		zones = append(zones, map[string]any{
			"name": fmt.Sprintf("Zone %d", i),
			"points": []map[string]float64{
				{"x": 0, "y": 0}, {"x": 1, "y": 0}, {"x": 1, "y": 1},
			},
		})
	}

	payload, err := json.Marshal(map[string]any{"walls": walls, "zones": zones})
	require.NoError(t, err)
	return payload
}

// duplicateIDs returns every id that appears more than once in a plan.
func duplicateIDs(plan domain.Plan) []string {
	seen := make(map[string]int)
	for _, wall := range plan.Walls {
		seen[wall.ID]++
		for _, opening := range wall.Openings {
			seen[opening.ID]++
		}
	}
	for _, zone := range plan.Zones {
		seen[zone.ID]++
	}

	var duplicates []string
	for id, count := range seen {
		if count > 1 {
			duplicates = append(duplicates, fmt.Sprintf("%s x%d", id, count))
		}
	}
	return duplicates
}

// TestConcurrentImportsProduceUniqueElementIDs is the regression test for the audit finding
// "Unsynchronised package-level idSeq in the .sh3d / AI-plan importers": the plan editor keys
// walls, openings and zones by id, so an import must never emit the same id twice, which is what
// happened when two requests imported plans concurrently and one rewound the shared counter.
func TestConcurrentImportsProduceUniqueElementIDs(t *testing.T) {
	payload := planFixture(t)
	const workers = 32

	var wg sync.WaitGroup
	start := make(chan struct{})
	failures := make(chan string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			plan, err := aijson.FromJSON(payload, "level-1")
			if err != nil {
				failures <- err.Error()
				return
			}
			if duplicates := duplicateIDs(plan); len(duplicates) > 0 {
				failures <- fmt.Sprintf("imported plan contains duplicate element ids: %v", duplicates)
			}
		}()
	}

	close(start)
	wg.Wait()
	close(failures)

	for failure := range failures {
		t.Error(failure)
	}
}

