package homeassistant

import "errors"

// ErrHomeAssistantUnavailable is returned by FetchStates when Home Assistant is
// unconfigured or unreachable. Callers that must not serve placeholder data — the
// background device poller — tell it apart from a successful fetch that legitimately
// returns an empty list.
var ErrHomeAssistantUnavailable = errors.New("home assistant is unavailable")
