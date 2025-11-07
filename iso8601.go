// Package iso8601 handles ISO8601-formatted durations.
package iso8601

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"time"
)

// Duration represents an ISO8601 Duration
// https://en.wikipedia.org/wiki/ISO_8601#Durations
type Duration struct {
	Y int
	M int
	W int
	D int
	// Time Component
	TH int
	TM int
	TS float64 // Seconds, can include fractional part (e.g., 33.3444)
}

var pattern = regexp.MustCompile(
	`^(-)?P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?` +
		`(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+(?:\.\d+)?)S)?)?$`)

// ParseISO8601 parses an ISO8601 duration string.
// Supports negative durations with a leading minus sign (e.g., -P1D).
//
//nolint:gocyclo // Complex parsing logic is necessary for ISO8601 format
func ParseISO8601(from string) (Duration, error) {
	var match []string
	var d Duration
	negative := false

	if pattern.MatchString(from) {
		match = pattern.FindStringSubmatch(from)
	} else {
		return d, errors.New("could not parse duration string")
	}

	// Check if the string starts with a minus sign
	if from != "" && from[0] == '-' {
		negative = true
	}

	for i, name := range pattern.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		switch name {
		case "year", "month", "week", "day", "hour", "minute": //nolint:goconst // These are field names, not constants
			val, err := strconv.Atoi(part)
			if err != nil {
				return d, err
			}
			if negative {
				val = -val
			}
			switch name {
			case "year":
				d.Y = val
			case "month":
				d.M = val
			case "week":
				d.W = val
			case "day":
				d.D = val
			case "hour":
				d.TH = val
			case "minute":
				d.TM = val
			}
		case "second":
			val, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return d, err
			}
			if negative {
				val = -val
			}
			d.TS = val
		}
	}

	return d, nil
}

// IsZero reports whether d represents the zero duration, P0D.
func (d Duration) IsZero() bool {
	return d.Y == 0 && d.M == 0 && d.W == 0 && d.D == 0 && d.TH == 0 && d.TM == 0 && d.TS == 0.0
}

// IsNegative returns true if the duration is negative.
func (d Duration) IsNegative() bool {
	return d.Y < 0 || d.M < 0 || d.W < 0 || d.D < 0 || d.TH < 0 || d.TM < 0 || d.TS < 0
}

// Negate returns a new Duration with all components negated.
func (d Duration) Negate() Duration {
	return Duration{
		Y:  -d.Y,
		M:  -d.M,
		W:  -d.W,
		D:  -d.D,
		TH: -d.TH,
		TM: -d.TM,
		TS: -d.TS,
	}
}

// HasTimePart returns true if the time part of the duration is non-zero.
func (d Duration) HasTimePart() bool {
	return d.TH != 0 || d.TM != 0 || d.TS != 0.0
}

// Shift returns a time.Time, shifted by the duration from the given start.
//
// NB: Shift uses time.AddDate for years, months, weeks, and days, and so
// shares its limitations. In particular, shifting by months is not recommended
// unless the start date is before the 28th of the month. Otherwise, dates will
// roll over, e.g. Aug 31 + P1M = Oct 1.
//
// Week and Day values will be combined as W*7 + D.
func (d Duration) Shift(t time.Time) time.Time {
	if d.Y != 0 || d.M != 0 || d.W != 0 || d.D != 0 {
		days := d.W*7 + d.D
		t = t.AddDate(d.Y, d.M, days)
	}
	t = t.Add(d.timeDuration())
	return t
}

// Unshift returns a time.Time, shifted back by the duration from the given start.
//
// NB: UnShift uses time.AddDate for years, months, weeks, and days, and so
// shares its limitations. In particular, shifting back by months is not recommended
// unless the start date is before the 28th of the month. Otherwise, dates will
// roll over, e.g. Oct 1 - P1M = Aug 31.
//
// Week and Day values will be combined as W*7 + D.
func (d Duration) Unshift(t time.Time) time.Time {
	if d.Y != 0 || d.M != 0 || d.W != 0 || d.D != 0 {
		days := d.W*7 + d.D
		t = t.AddDate(-d.Y, -d.M, -days)
	}
	t = t.Add(-d.timeDuration())
	return t
}

func (d Duration) timeDuration() time.Duration {
	var dur time.Duration
	dur += time.Duration(d.TH) * time.Hour
	dur += time.Duration(d.TM) * time.Minute
	// Convert fractional seconds to nanoseconds
	dur += time.Duration(d.TS * float64(time.Second))
	return dur
}

var tmpl = template.Must(template.New("duration").Funcs(template.FuncMap{
	"formatSeconds": func(ts float64) string {
		if ts == float64(int64(ts)) {
			return fmt.Sprintf("%.0f", ts)
		}
		// Remove trailing zeros
		return fmt.Sprintf("%g", ts)
	},
	"isNegative": func(d Duration) bool {
		return d.Y < 0 || d.M < 0 || d.W < 0 || d.D < 0 || d.TH < 0 || d.TM < 0 || d.TS < 0
	},
	"abs": func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	},
	"absFloat": func(f float64) float64 {
		if f < 0 {
			return -f
		}
		return f
	},
}).Parse(
	`{{if isNegative .}}-{{end}}P{{if .Y}}{{abs .Y}}Y{{end}}{{if .M}}{{abs .M}}M{{end}}` +
		`{{if .W}}{{abs .W}}W{{end}}{{if .D}}{{abs .D}}D{{end}}{{if .HasTimePart}}T{{end }}` +
		`{{if .TH}}{{abs .TH}}H{{end}}{{if .TM}}{{abs .TM}}M{{end}}{{if .TS}}{{formatSeconds (absFloat .TS)}}S{{end}}`))

// String returns an ISO8601-ish representation of the duration.
func (d Duration) String() string {
	var s bytes.Buffer

	if d.IsZero() {
		return "P0D"
	}

	err := tmpl.Execute(&s, d)
	if err != nil {
		panic(err)
	}

	return s.String()
}

// MarshalJSON satisfies json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON satisfies json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, err := ParseISO8601(s)
	if err != nil {
		return err
	}
	*d = tmp

	return nil
}

// Add returns a new Duration that is the sum of d and other.
// Note: This performs component-wise addition. For durations with months/years,
// the result may not represent the exact calendar duration due to variable month lengths.
func (d Duration) Add(other Duration) Duration {
	return Duration{
		Y:  d.Y + other.Y,
		M:  d.M + other.M,
		W:  d.W + other.W,
		D:  d.D + other.D,
		TH: d.TH + other.TH,
		TM: d.TM + other.TM,
		TS: d.TS + other.TS,
	}
}

// Subtract returns a new Duration that is the difference of d and other.
// Note: This performs component-wise subtraction. For durations with months/years,
// the result may not represent the exact calendar duration due to variable month lengths.
func (d Duration) Subtract(other Duration) Duration {
	return Duration{
		Y:  d.Y - other.Y,
		M:  d.M - other.M,
		W:  d.W - other.W,
		D:  d.D - other.D,
		TH: d.TH - other.TH,
		TM: d.TM - other.TM,
		TS: d.TS - other.TS,
	}
}

// Multiply returns a new Duration with all components multiplied by n.
func (d Duration) Multiply(n int) Duration {
	return Duration{
		Y:  d.Y * n,
		M:  d.M * n,
		W:  d.W * n,
		D:  d.D * n,
		TH: d.TH * n,
		TM: d.TM * n,
		TS: d.TS * float64(n),
	}
}

// Equal returns true if d and other have identical components.
// Note: For durations with months/years, equal components may represent different
// calendar durations due to variable month lengths.
func (d Duration) Equal(other Duration) bool {
	return d.Y == other.Y && d.M == other.M && d.W == other.W && d.D == other.D &&
		d.TH == other.TH && d.TM == other.TM && d.TS == other.TS
}

// LessThan returns true if d is less than other.
// This comparison is only meaningful for time-only durations (no years/months/weeks/days).
// For durations with date components, the result may be ambiguous due to variable month lengths.
func (d Duration) LessThan(other Duration) bool {
	// If either has date components, comparison is ambiguous
	if d.Y != 0 || d.M != 0 || d.W != 0 || d.D != 0 ||
		other.Y != 0 || other.M != 0 || other.W != 0 || other.D != 0 {
		// Fall back to component-wise comparison for date parts
		if d.Y != other.Y {
			return d.Y < other.Y
		}
		if d.M != other.M {
			return d.M < other.M
		}
		if d.W != other.W {
			return d.W < other.W
		}
		if d.D != other.D {
			return d.D < other.D
		}
	}
	// Compare time components
	if d.TH != other.TH {
		return d.TH < other.TH
	}
	if d.TM != other.TM {
		return d.TM < other.TM
	}
	return d.TS < other.TS
}

// GreaterThan returns true if d is greater than other.
// This comparison is only meaningful for time-only durations (no years/months/weeks/days).
// For durations with date components, the result may be ambiguous due to variable month lengths.
func (d Duration) GreaterThan(other Duration) bool {
	return other.LessThan(d)
}

// ToTimeDuration converts the time component of d to a time.Duration.
// Date components (years, months, weeks, days) are ignored.
func (d Duration) ToTimeDuration() time.Duration {
	return d.timeDuration()
}

// FromTimeDuration creates a Duration from a time.Duration.
// Only the time component is set; date components are zero.
func FromTimeDuration(td time.Duration) Duration {
	d := Duration{}
	hours := td / time.Hour
	td -= hours * time.Hour
	minutes := td / time.Minute
	td -= minutes * time.Minute
	seconds := float64(td) / float64(time.Second)

	d.TH = int(hours)
	d.TM = int(minutes)
	d.TS = seconds

	return d
}
