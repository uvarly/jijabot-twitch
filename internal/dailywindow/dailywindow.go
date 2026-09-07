package dailywindow

import "time"

func Key(t time.Time, resetHour int) string {
	shifted := t.UTC().Add(-time.Duration(resetHour) * time.Hour)
	return shifted.Format("2006-01-02")
}
