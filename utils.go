package guildrone

import "time"

func parseTime(raw string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC3339, raw)
	return
}
