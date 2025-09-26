package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type ParseStringSupportTypes interface {
	string |
		int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float64 |
		bool | time.Duration | url.URL
}

// ParseString attempts to parse the input string `s` into a value of the specified type T.
// If parsing the string `s` fails for a supported type, it returns the zero value of T
// and the parsing error.
func ParseString[T ParseStringSupportTypes](rawValue string) (T, error) {
	var value T

	switch any(value).(type) {
	case int:
		i, err := strconv.Atoi(rawValue)
		if err != nil {
			return value, err
		}
		value = any(i).(T)
	case int8:
		i, err := strconv.ParseInt(rawValue, 10, 8)
		if err != nil {
			return value, err
		}
		value = any(int8(i)).(T)
	case int16:
		i, err := strconv.ParseInt(rawValue, 10, 16)
		if err != nil {
			return value, err
		}
		value = any(int16(i)).(T)
	case int32:
		i, err := strconv.ParseInt(rawValue, 10, 32)
		if err != nil {
			return value, err
		}
		value = any(int32(i)).(T)
	case int64:
		i, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return value, err
		}
		value = any(i).(T)
	case uint:
		u, err := strconv.ParseUint(rawValue, 10, 0)
		if err != nil {
			return value, err
		}
		value = any(uint(u)).(T)
	case uint8:
		u, err := strconv.ParseUint(rawValue, 10, 8)
		if err != nil {
			return value, err
		}
		value = any(uint8(u)).(T)
	case uint16:
		u, err := strconv.ParseUint(rawValue, 10, 16)
		if err != nil {
			return value, err
		}
		value = any(uint16(u)).(T)
	case uint32:
		u, err := strconv.ParseUint(rawValue, 10, 32)
		if err != nil {
			return value, err
		}
		value = any(uint32(u)).(T)
	case uint64:
		u, err := strconv.ParseUint(rawValue, 10, 64)
		if err != nil {
			return value, err
		}
		value = any(u).(T)
	case float64:
		f, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return value, err
		}
		value = any(f).(T)
	case bool:
		b, err := strconv.ParseBool(rawValue)
		if err != nil {
			return value, err
		}
		value = any(b).(T)
	case string:
		value = any(rawValue).(T)
	case time.Duration:
		d, err := time.ParseDuration(rawValue)
		if err != nil {
			return value, err
		}
		value = any(d).(T)
	case url.URL:
		u, err := url.Parse(rawValue)
		if err != nil {
			return value, err
		}
		value = any(*u).(T)
	default:
		return value, fmt.Errorf("unsupported type: %T", value)
	}

	return value, nil
}
