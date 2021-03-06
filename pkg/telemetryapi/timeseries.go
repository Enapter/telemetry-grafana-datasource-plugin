package telemetryapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Timeseries struct {
	TimeField  []time.Time
	DataFields []*TimeseriesDataField
}

func (ts *Timeseries) Len() int {
	return len(ts.TimeField)
}

type TimeseriesTags map[string]string

func parseTimeseriesTags(s string) (TimeseriesTags, error) {
	tags := make(TimeseriesTags)

	pairs := strings.Split(s, " ")

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")

		if want := 2; len(kv) != want {
			return nil, fmt.Errorf("%w: len: want %d, have %d",
				errBadKeyValuePair, want, len(kv))
		}

		tags[kv[0]] = kv[1]
	}

	return tags, nil
}

type TimeseriesDataField struct {
	Tags   TimeseriesTags
	Type   TimeseriesDataType
	Values []interface{}
}

type TimeseriesDataType uint8

const (
	TimeseriesDataTypeUnknown = iota
	TimeseriesDataTypeFloat64
	TimeseriesDataTypeInt64
	TimeseriesDataTypeString
	TimeseriesDataTypeStringArray
	TimeseriesDataTypeBool
)

func (t TimeseriesDataType) ZeroValue() interface{} {
	switch t {
	case TimeseriesDataTypeFloat64:
		return (*float64)(nil)
	case TimeseriesDataTypeInt64:
		return (*int64)(nil)
	case TimeseriesDataTypeString:
		return (*string)(nil)
	case TimeseriesDataTypeStringArray:
		return ([]string)(nil)
	case TimeseriesDataTypeBool:
		return (*bool)(nil)
	default:
		return nil
	}
}

func parseTimeseriesDataTypes(ss []string) ([]TimeseriesDataType, error) {
	dataTypes := make([]TimeseriesDataType, len(ss))

	for i, s := range ss {
		t, err := parseTimeseriesDataType(s)
		if err != nil {
			return nil, fmt.Errorf("%d: %w", i, err)
		}
		dataTypes[i] = t
	}

	return dataTypes, nil
}

func parseTimeseriesDataType(s string) (TimeseriesDataType, error) {
	switch s {
	case "float64":
		return TimeseriesDataTypeFloat64, nil
	case "int64":
		return TimeseriesDataTypeInt64, nil
	case "string":
		return TimeseriesDataTypeString, nil
	case "[]string":
		return TimeseriesDataTypeStringArray, nil
	case "bool":
		return TimeseriesDataTypeBool, nil
	default:
		return TimeseriesDataTypeUnknown, fmt.Errorf("%w: %s",
			errUnexpectedTimeseriesDataType, s)
	}
}

func (t TimeseriesDataType) String() string {
	switch t {
	case TimeseriesDataTypeFloat64:
		return "float64"
	case TimeseriesDataTypeInt64:
		return "int64"
	case TimeseriesDataTypeString:
		return "string"
	case TimeseriesDataTypeStringArray:
		return "[]string"
	case TimeseriesDataTypeBool:
		return "bool"
	default:
		return "unknown"
	}
}

func (t TimeseriesDataType) Parse(s string) (interface{}, error) {
	if len(s) == 0 {
		return t.ZeroValue(), nil
	}

	switch t {
	case TimeseriesDataTypeFloat64:
		const bitSize = 64
		v, err := strconv.ParseFloat(s, bitSize)
		if err != nil {
			return nil, err
		}
		return &v, nil
	case TimeseriesDataTypeInt64:
		const base = 10
		const bitSize = 64
		v, err := strconv.ParseInt(s, base, bitSize)
		if err != nil {
			return nil, err
		}
		return &v, nil
	case TimeseriesDataTypeString:
		return &s, nil
	case TimeseriesDataTypeStringArray:
		var values []string
		if err := json.Unmarshal([]byte(s), &values); err != nil {
			return nil, err
		}
		return values, nil
	case TimeseriesDataTypeBool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, errUnexpectedTimeseriesDataType
	}
}

func NewTimeseries(dataTypes []TimeseriesDataType) *Timeseries {
	const preallocValues = 64

	timeField := make([]time.Time, 0, preallocValues)

	dataFields := make([]*TimeseriesDataField, len(dataTypes))
	for i, dataType := range dataTypes {
		dataFields[i] = &TimeseriesDataField{
			Tags:   make(map[string]string),
			Type:   dataType,
			Values: make([]interface{}, 0, preallocValues),
		}
	}

	return &Timeseries{
		TimeField:  timeField,
		DataFields: dataFields,
	}
}
