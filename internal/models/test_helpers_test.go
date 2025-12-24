package models

import "time"

func testPastDateUTC() time.Time {
	return time.Now().UTC().AddDate(0, 0, -30)
}
