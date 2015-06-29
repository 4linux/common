package model

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// TODO(fabxc): these types are to be migrated over from Prometheus.

type ValueType int

const (
	ValNone ValueType = iota
	ValScalar
	ValVector
	ValMatrix
	ValString
)

// MarshalJSON implements json.Marshaler.
func (et ValueType) MarshalJSON() ([]byte, error) {
	return json.Marshal(et.String())
}

func (et *ValueType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "<ValNone>":
		*et = ValNone
	case "scalar":
		*et = ValScalar
	case "vector":
		*et = ValVector
	case "matrix":
		*et = ValMatrix
	case "string":
		*et = ValString
	default:
		return fmt.Errorf("unknown value type %q", s)
	}
	return nil
}

func (e ValueType) String() string {
	switch e {
	case ValNone:
		return "<ValNone>"
	case ValScalar:
		return "scalar"
	case ValVector:
		return "vector"
	case ValMatrix:
		return "matrix"
	case ValString:
		return "string"
	}
	panic("ValueType.String: unhandled value type")
}

// SampleStream is a stream of Values belonging to an attached COWMetric.
type SampleStream struct {
	Metric Metric       `json:"metric"`
	Values []SamplePair `json:"values"`
}

// // Sample is a single sample belonging to a COWMetric.
// type Sample struct {
// 	Metric    COWMetric   `json:"metric"`
// 	Value     SampleValue `json:"value"`
// 	Timestamp Timestamp   `json:"timestamp"`
// }

// Scalar is a scalar value evaluated at the set timestamp.
type Scalar struct {
	Value     SampleValue `json:"value"`
	Timestamp Time        `json:"timestamp"`
}

func (s *Scalar) String() string {
	return fmt.Sprintf("scalar: %v @[%v]", s.Value, s.Timestamp)
}

// String is a string value evaluated at the set timestamp.
type String struct {
	Value     string `json:"value"`
	Timestamp Time   `json:"timestamp"`
}

func (s *String) String() string {
	return s.Value
}

// Vector is basically only an alias for Samples, but the
// contract is that in a Vector, all Samples have the same timestamp.
type Vector []*Sample

// Matrix is a slice of SampleStreams that implements sort.Interface and
// has a String method.
type Matrix []*SampleStream

// Len implements sort.Interface.
func (m Matrix) Len() int {
	return len(m)
}

// Less implements sort.Interface.
func (m Matrix) Less(i, j int) bool {
	return m[i].Metric.String() < m[j].Metric.String()
}

// Swap implements sort.Interface.
func (m Matrix) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// Value is a generic interface for values resulting from a query evaluation.
type Value interface {
	Type() ValueType
	String() string
}

func (matrix Matrix) String() string {
	metricStrings := make([]string, 0, len(matrix))
	for _, sampleStream := range matrix {
		metricName, hasName := sampleStream.Metric.Has(MetricNameLabel)
		numLabels := sampleStream.Metric.Len()
		if hasName {
			numLabels--
		}
		labelStrings := make([]string, 0, numLabels)
		for label, value := range sampleStream.Metric.LabelSet {
			if label != MetricNameLabel {
				labelStrings = append(labelStrings, fmt.Sprintf("%s=%q", label, value))
			}
		}
		sort.Strings(labelStrings)
		valueStrings := make([]string, 0, len(sampleStream.Values))
		for _, value := range sampleStream.Values {
			valueStrings = append(valueStrings,
				fmt.Sprintf("\n%v @[%v]", value.Value, value.Timestamp))
		}
		metricStrings = append(metricStrings,
			fmt.Sprintf("%s{%s} => %s",
				metricName,
				strings.Join(labelStrings, ", "),
				strings.Join(valueStrings, ", ")))
	}
	sort.Strings(metricStrings)
	return strings.Join(metricStrings, "\n")
}

func (vector Vector) String() string {
	metricStrings := make([]string, 0, len(vector))
	for _, sample := range vector {
		metricStrings = append(metricStrings,
			fmt.Sprintf("%s => %v @[%v]",
				sample.Metric,
				sample.Value, sample.Timestamp))
	}
	return strings.Join(metricStrings, "\n")
}

func (Matrix) Type() ValueType  { return ValMatrix }
func (Vector) Type() ValueType  { return ValVector }
func (*Scalar) Type() ValueType { return ValScalar }
func (*String) Type() ValueType { return ValString }

// Interval describes the inclusive interval between two Timestamps.
type Interval struct {
	OldestInclusive Time
	NewestInclusive Time
}
