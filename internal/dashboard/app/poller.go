package app

import (
	"context"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/pkg/logger"
)

// DefaultDevicePollInterval is how often device states are refreshed when no interval is configured.
const DefaultDevicePollInterval = 10 * time.Second

// devicePollTimeout caps a single Home Assistant fetch so a hung request cannot stall
// every subsequent poll.
const devicePollTimeout = 20 * time.Second

// DeviceStateSource fetches the current state of every supported Home Assistant device.
// Unlike the REST-facing repository it must return an error (never placeholder data) when
// the instance is unreachable, so the poller never broadcasts mock states.
type DeviceStateSource interface {
	FetchStates(ctx context.Context) ([]domain.Device, error)
}

// DevicePoller watches Home Assistant device states in the background and broadcasts every
// change — including a device dropping to `unavailable` when it loses the network — to the
// WebSocket hub, so the dashboard's offline indicator reflects reality without a page reload.
type DevicePoller struct {
	source      DeviceStateSource
	broadcaster domain.DeviceBroadcaster
	interval    time.Duration
}

// NewDevicePoller creates a poller. A non-positive interval selects DefaultDevicePollInterval.
func NewDevicePoller(source DeviceStateSource, broadcaster domain.DeviceBroadcaster, interval time.Duration) *DevicePoller {
	if interval <= 0 {
		interval = DefaultDevicePollInterval
	}
	return &DevicePoller{source: source, broadcaster: broadcaster, interval: interval}
}

// Run polls until ctx is cancelled. The first successful poll only seeds the snapshot:
// clients already fetched the current states over REST, so echoing every device back would
// be noise. Every later change is broadcast.
func (p *DevicePoller) Run(ctx context.Context) {
	log := logger.LoggerFromContext(ctx)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	known := make(map[string]domain.Device)
	seeded := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		devices, err := p.fetch(ctx)
		if err != nil {
			// Keep the previous snapshot: a Home Assistant outage must not be mistaken
			// for every device going offline.
			log.WithError(err).Debug("device state poll failed, keeping last known states")
			continue
		}

		for _, device := range devices {
			previous, seen := known[device.ID]
			known[device.ID] = device
			if !seeded {
				continue
			}
			if !seen || deviceStateChanged(previous, device) {
				if p.broadcaster != nil {
					p.broadcaster.BroadcastDevice(device)
				}
			}
		}

		if !seeded {
			log.WithField("devices", len(known)).Info("device state poller seeded")
			seeded = true
		}
	}
}

func (p *DevicePoller) fetch(ctx context.Context) ([]domain.Device, error) {
	fetchCtx, cancel := context.WithTimeout(ctx, devicePollTimeout)
	defer cancel()
	return p.source.FetchStates(fetchCtx)
}

// deviceStateChanged reports whether a device must be re-broadcast. Home Assistant bumps
// last_updated on every state or attribute write, so comparing the state and the timestamp
// covers value, brightness and climate changes without a deep attribute diff.
func deviceStateChanged(previous, current domain.Device) bool {
	return previous.State != current.State || !previous.LastUpdated.Equal(current.LastUpdated)
}
