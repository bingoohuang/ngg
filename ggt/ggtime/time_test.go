package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ExampleInternalTime() {
	var jt jsonType

	json1 := `{
	  "num_seconds": 1651808102,
	  "num_milliseconds": 1651808102363,
	  "num_microseconds": 1651808102363368,
	  "num_nanoseconds": 1651808102363368423,

	  "hex_seconds": "0x62749766",

	  "str_seconds": "1651808102",
	  "str_milliseconds": "1651808102363",
	  "str_microseconds": "1651808102363368",
	  "str_nanoseconds": "1651808102363368423",

	  "str_rfc3339": "2022-05-06T03:35:02Z",
	  "str_rfc3339_nano": "2022-05-06T03:35:02.363368423Z",
	  "str_rfc1123": "Fri, 06 May 2022 03:35:02 UTC",
	  "str_rfc850": "Friday, 06-May-22 03:35:02 UTC",
	  "str_rfc822": "06 May 22 03:35 UTC"
	}`

	err := json.Unmarshal([]byte(json1), &jt)
	if err != nil {
		panic(err)
	}

	json2, _ := json.MarshalIndent(jt, "", "  ")

	fmt.Println(string(json2))
	// Output:
	// {
	//   "num_seconds": "2022-05-06T11:35:02+08:00",
	//   "num_milliseconds": "2022-05-06T11:35:02.363+08:00",
	//   "num_microseconds": "2022-05-06T11:35:02.363368+08:00",
	//   "num_nanoseconds": "2022-05-06T11:35:02.363368423+08:00",
	//   "hex_seconds": "2022-05-06T11:35:02+08:00",
	//   "str_seconds": "2022-05-06T11:35:02+08:00",
	//   "str_milliseconds": "2022-05-06T11:35:02.363+08:00",
	//   "str_microseconds": "2022-05-06T11:35:02.363368+08:00",
	//   "str_nanoseconds": "2022-05-06T11:35:02.363368423+08:00",
	//   "str_rfc3339": "2022-05-06T03:35:02Z",
	//   "str_rfc3339_nano": "2022-05-06T03:35:02.363368423Z",
	//   "str_rfc1123": "2022-05-06T03:35:02Z",
	//   "str_rfc850": "2022-05-06T03:35:02Z",
	//   "str_rfc822": "2022-05-06T03:35:00Z"
	// }
}

type jsonType struct {
	NumSeconds      InternalTime `json:"num_seconds"`
	NumMilliseconds InternalTime `json:"num_milliseconds"`
	NumMicroseconds InternalTime `json:"num_microseconds"`
	NumNanoseconds  InternalTime `json:"num_nanoseconds"`

	HexSeconds InternalTime `json:"hex_seconds"`

	StrSeconds      InternalTime `json:"str_seconds"`
	StrMilliseconds InternalTime `json:"str_milliseconds"`
	StrMicroseconds InternalTime `json:"str_microseconds"`
	StrNanoseconds  InternalTime `json:"str_nanoseconds"`

	StrRFC3339     InternalTime `json:"str_rfc3339"`
	StrRFC3339Nano InternalTime `json:"str_rfc3339_nano"`
	StrRFC1123     InternalTime `json:"str_rfc1123"`
	StrRFC850      InternalTime `json:"str_rfc850"`
	StrRFC822      InternalTime `json:"str_rfc822"`
}

type InternalTime struct {
	time.Time
}

func (it *InternalTime) UnmarshalJSON(data []byte) error {
	// Make sure that the input is not empty
	if len(data) == 0 {
		return errors.New("empty value is not supported")
	}

	// If the input is not a string, try to parse it as a number, otherwise return an error.
	if data[0] != '"' {
		timeInt64, err := strconv.ParseInt(string(data), 0, 64)
		if err != nil {
			return err
		}

		it.Time = parseTimestamp(timeInt64)
	}

	// If the input is a string, trim quotes.
	str := strings.Trim(string(data), `"`)

	// Parse the string as a time using the supported layouts.
	parsed, err := parseTime(formats, str)
	if err == nil {
		it.Time = parsed

		return nil
	}

	// As the final attempt, try to parse the string as a timestamp.
	timeInt64, err := strconv.ParseInt(str, 0, 64)
	if err == nil {
		it.Time = parseTimestamp(timeInt64)

		return nil
	}

	return errors.New("unknown time format")
}

func parseTimestamp(timestamp int64) time.Time {
	switch {
	case timestamp < minMicroseconds:
		return time.Unix(0, timestamp) // Before 1970 in nanoseconds.
	case timestamp < minMilliseconds:
		return time.Unix(0, timestamp*int64(time.Microsecond)) // Before 1970 in microseconds.
	case timestamp < minSeconds:
		return time.Unix(0, timestamp*int64(time.Millisecond)) // Before 1970 in milliseconds.
	case timestamp < 0:
		return time.Unix(timestamp, 0) // Before 1970 in seconds.
	case timestamp < maxSeconds:
		return time.Unix(timestamp, 0) // After 1970 in seconds.
	case timestamp < maxMilliseconds:
		return time.Unix(0, timestamp*int64(time.Millisecond)) // After 1970 in milliseconds.
	case timestamp < maxMicroseconds:
		return time.Unix(0, timestamp*int64(time.Microsecond)) // After 1970 in microseconds.
	}

	return time.Unix(0, timestamp) // After 1970 in nanoseconds.
}

func parseTime(formats []string, dt string) (time.Time, error) {
	for _, format := range formats {
		parsedTime, err := time.Parse(format, dt)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", dt)
}
