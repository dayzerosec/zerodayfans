package commafeed_util

import "time"

type TimePrimitive float32

// PrimitiveToTime will convert the time from Commafeed to a time object, annoyingly the OpenAPI implementation keeps
// changing between a string and a number so this function can quickly change to handle either
func PrimitiveToTime(date TimePrimitive) (time.Time, error) {
	return float32ToTime(float32(date))
	//return stringToTime(string(date))
}

func float32ToTime(date float32) (time.Time, error) {
	return time.UnixMilli(int64(date)), nil
}

func stringToTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)
}
