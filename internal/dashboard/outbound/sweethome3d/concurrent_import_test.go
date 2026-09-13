package sweethome3d_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sweethome3d"
	"github.com/stretchr/testify/require"
)

// sh3dFixture builds a small .sh3d archive with two levels whose walls carry openings and rooms,
// so a single import derives dozens of element ids.
func sh3dFixture(t testing.TB) []byte {
	t.Helper()
	var xml strings.Builder
	xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?><home>`)
	xml.WriteString(`<level id="l0" name="Ground" elevation="0"/>`)
	xml.WriteString(`<level id="l1" name="Upper" elevation="300"/>`)
	for _, level := range []string{"l0", "l1"} {
		for i := 0; i < 30; i++ {
			fmt.Fprintf(&xml, `<wall id="%s-w%d" xStart="%d" yStart="0" xEnd="%d" yEnd="0" thickness="10" level="%s"/>`,
				level, i, i*100, i*100+80, level)
			fmt.Fprintf(&xml, `<doorOrWindow wall="%s-w%d" x="%d" y="0" width="80" name="door" level="%s"/>`,
				level, i, i*100+40, level)
		}
		for i := 0; i < 5; i++ {
			fmt.Fprintf(&xml, `<room name="Room %s %d" level="%s"><point x="0" y="0"/><point x="100" y="0"/><point x="100" y="100"/></room>`,
				level, i, level)
		}
	}
	xml.WriteString(`</home>`)

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	entry, err := writer.Create("Home.xml")
	require.NoError(t, err)
	_, err = entry.Write([]byte(xml.String()))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return buf.Bytes()
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
// "Unsynchronised package-level idSeq in the .sh3d / AI-plan importers". Element ids are the
// identity the plan editor keys on, so an import must never emit a plan that contains the same id
// twice — which is what happened when two HTTP requests imported plans at the same time and one
// rewound the shared counter the other was using. The race detector turns the same defect into a
// hard failure.
func TestConcurrentImportsProduceUniqueElementIDs(t *testing.T) {
	archive := sh3dFixture(t)
	const workers = 32

	var wg sync.WaitGroup
	start := make(chan struct{})
	failures := make(chan string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			levels, err := sweethome3d.FromReaderLevels(bytes.NewReader(archive))
			if err != nil {
				failures <- err.Error()
				return
			}
			for _, level := range levels {
				if duplicates := duplicateIDs(level.Plan); len(duplicates) > 0 {
					failures <- fmt.Sprintf("imported plan %q contains duplicate element ids: %v", level.Name, duplicates)
					return
				}
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
