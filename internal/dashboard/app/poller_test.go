package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scriptedStateSource struct {
	mu    sync.Mutex
	calls int
	next  func(call int) ([]domain.Device, error)
}

func (s *scriptedStateSource) FetchStates(ctx context.Context) ([]domain.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	call := s.calls
	s.calls++
	if s.next == nil {
		return nil, nil
	}
	return s.next(call)
}

func (s *scriptedStateSource) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

type recordingBroadcaster struct {
	mu      sync.Mutex
	devices []domain.Device
}

func (r *recordingBroadcaster) BroadcastDevice(device domain.Device) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices = append(r.devices, device)
}

func (r *recordingBroadcaster) snapshot() []domain.Device {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Device, len(r.devices))
	copy(out, r.devices)
	return out
}

func kitchenLamp(state string, updated time.Time) domain.Device {
	return domain.Device{
		ID:          "light.kitchen",
		Name:        "Kitchen lamp",
		Domain:      domain.DomainLight,
		State:       state,
		Attributes:  map[string]any{},
		LastUpdated: updated,
	}
}

func TestDevicePoller_SeedsThenBroadcastsOnlyChanges(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	source := &scriptedStateSource{next: func(call int) ([]domain.Device, error) {
		switch call {
		case 0:
			return []domain.Device{kitchenLamp("on", base)}, nil
		case 1:
			// Unchanged since the seed: must stay silent.
			return []domain.Device{kitchenLamp("on", base)}, nil
		default:
			// The lamp dropped off the network.
			return []domain.Device{kitchenLamp("unavailable", base.Add(time.Second))}, nil
		}
	}}
	recorder := &recordingBroadcaster{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go NewDevicePoller(source, recorder, 5*time.Millisecond).Run(ctx)

	require.Eventually(t, func() bool {
		got := recorder.snapshot()
		return len(got) == 1 && got[0].State == "unavailable"
	}, time.Second, 5*time.Millisecond)

	assert.GreaterOrEqual(t, source.callCount(), 3)
	// The transition is broadcast once; identical follow-up polls stay silent.
	time.Sleep(30 * time.Millisecond)
	assert.Len(t, recorder.snapshot(), 1)
}

func TestDevicePoller_SkipsBroadcastWhenFetchFails(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	source := &scriptedStateSource{next: func(call int) ([]domain.Device, error) {
		if call == 0 {
			return []domain.Device{kitchenLamp("on", base)}, nil
		}
		return nil, errors.New("HA unreachable")
	}}
	recorder := &recordingBroadcaster{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go NewDevicePoller(source, recorder, 5*time.Millisecond).Run(ctx)

	assert.Never(t, func() bool { return len(recorder.snapshot()) > 0 }, 50*time.Millisecond, 5*time.Millisecond)
}

func TestDevicePoller_BroadcastsNewlyAppearedDevice(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	source := &scriptedStateSource{next: func(call int) ([]domain.Device, error) {
		if call == 0 {
			return []domain.Device{kitchenLamp("on", base)}, nil
		}
		plug := domain.Device{
			ID:          "switch.plug",
			Name:        "Plug",
			Domain:      domain.DomainSwitch,
			State:       "on",
			Attributes:  map[string]any{},
			LastUpdated: base,
		}
		return []domain.Device{kitchenLamp("on", base), plug}, nil
	}}
	recorder := &recordingBroadcaster{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go NewDevicePoller(source, recorder, 5*time.Millisecond).Run(ctx)

	require.Eventually(t, func() bool {
		got := recorder.snapshot()
		return len(got) == 1 && got[0].ID == "switch.plug"
	}, time.Second, 5*time.Millisecond)
}
