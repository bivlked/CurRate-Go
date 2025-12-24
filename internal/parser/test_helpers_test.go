package parser

import "time"

func testPastDateUTC() time.Time {
	return time.Now().UTC().AddDate(0, 0, -30)
}

func formatCBRDate(date time.Time) string {
	return date.Format("02.01.2006")
}
