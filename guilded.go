package guildrone

import (
	"net/http"
	"time"
)

// VERSION of Guildrone, follows Semantic Versioning. (http://semver.org/)
const VERSION = "0.3.5"

// New creates a new Guilded session with provided token
func New(token string) (s *Session, err error) {
	s = &Session{
		Token:                         token,
		ShouldReconnectOnError:        true,
		ShouldReplayEventsOnReconnect: true,
		ShouldRetryOnRateLimit:        true,
		MaxRestRetries:                3,
		Client:                        &http.Client{Timeout: (20 * time.Second)},
		UserAgent:                     "GuildedBot (https://github.com/FlameInTheDark/guildrone, v" + VERSION + ")",
		LastHeartbeatAck:              time.Now().UTC(),
	}

	return
}
