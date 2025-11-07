package iso8601_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AbhijitDhariya/iso8601"
)

const dateLayout = "Jan 2, 2006 at 03:04:05"

func makeTime(t *testing.T, s string) time.Time {
	result, err := time.Parse(dateLayout, s)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCanShift(t *testing.T) {
	cases := []struct {
		from     string
		duration iso8601.Duration
		want     string
	}{
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{}, "Jan 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{Y: 1}, "Jan 1, 2019 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{M: 1}, "Feb 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{M: 2}, "Mar 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{W: 1}, "Jan 8, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{D: 1}, "Jan 2, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{TH: 1}, "Jan 1, 2018 at 01:00:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{TM: 1}, "Jan 1, 2018 at 00:01:00"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{TS: 1}, "Jan 1, 2018 at 00:00:01"},
		{"Jan 1, 2018 at 00:00:00", iso8601.Duration{
			Y:  10,
			M:  5,
			D:  8,
			TH: 5,
			TM: 10,
			TS: 6,
			//T: 5*time.Hour + 10*time.Minute + 6*time.Second,
		},
			"Jun 9, 2028 at 05:10:06",
		},
	}

	for k, c := range cases {
		from := makeTime(t, c.from)
		want := makeTime(t, c.want)

		got := c.duration.Shift(from)
		if !want.Equal(got) {
			t.Fatalf("Case %d: want=%s, got=%s", k, want, got)
		}
	}
}

func TestCanMaintainHourThroughDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	current, err := time.ParseInLocation(dateLayout, "Jan 1, 2018 at 00:00:00", loc)
	if err != nil {
		t.Fatal(err)
	}

	sut := iso8601.Duration{D: 1}
	for d := 0; d < 365; d++ {
		if got := current.Hour(); got != 0 {
			t.Fatalf("Day %d: want=0, got=%d", d, got)
		}
		current = sut.Shift(current)
	}
}

func TestCanParse(t *testing.T) {
	cases := []struct {
		from string
		want iso8601.Duration
	}{
		{"P1Y", iso8601.Duration{Y: 1}},
		{"P1M", iso8601.Duration{M: 1}},
		{"P2M", iso8601.Duration{M: 2}},
		{"P1W", iso8601.Duration{W: 1}},
		{"P1D", iso8601.Duration{D: 1}},
		{"PT1H", iso8601.Duration{TH: 1}},
		{"PT1M", iso8601.Duration{TM: 1}},
		{"PT1S", iso8601.Duration{TS: 1}},
		{"P10Y5M8DT5H10M6S", iso8601.Duration{Y: 10, M: 5, D: 8, TH: 5, TM: 10, TS: 6}},
	}

	for k, c := range cases {
		got, err := iso8601.ParseISO8601(c.from)
		if err != nil {
			t.Fatal(err)
		}
		if c.want != got {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestCanParseFractionalSeconds(t *testing.T) {
	cases := []struct {
		from string
		want iso8601.Duration
	}{
		{"PT33.3444S", iso8601.Duration{TS: 33.3444}},
		{"PT0.5S", iso8601.Duration{TS: 0.5}},
		{"PT1.123S", iso8601.Duration{TS: 1.123}},
		{"P343DT13H8M33.3444S", iso8601.Duration{D: 343, TH: 13, TM: 8, TS: 33.3444}},
		{"PT1.999999S", iso8601.Duration{TS: 1.999999}},
	}

	for k, c := range cases {
		got, err := iso8601.ParseISO8601(c.from)
		if err != nil {
			t.Fatalf("Case %d: failed to parse %s: %v", k, c.from, err)
		}
		// Compare with epsilon for float comparison
		if c.want.Y != got.Y || c.want.M != got.M || c.want.W != got.W || c.want.D != got.D ||
			c.want.TH != got.TH || c.want.TM != got.TM {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
		epsilon := 0.0001
		diff := c.want.TS - got.TS
		if diff < 0 {
			diff = -diff
		}
		if diff > epsilon {
			t.Fatalf("Case %d: TS want=%f, got=%f", k, c.want.TS, got.TS)
		}
	}
}

func TestCanRejectBadString(t *testing.T) {
	cases := []string{
		"",
		"PP1D",
		"P1D2F",
		"P2F",
	}

	for _, c := range cases {
		_, err := iso8601.ParseISO8601(c)
		if err == nil {
			t.Fatalf("%s: Expected error, got none", c)
		}
	}
}

func TestCanStringifyZeroValue(t *testing.T) {
	sut := iso8601.Duration{}
	want := "P0D"
	got := sut.String()
	if want != got {
		t.Fatalf("want=%s, got=%s", want, got)
	}
}

func TestCanStringify(t *testing.T) {
	cases := []string{
		"P1Y",
		"P2M",
		"P3W",
		"P4D",
		"PT5H",
		"PT6M",
		"PT7S",
		"P1Y2M3W4DT5H6M7S",
	}
	for _, want := range cases {
		sut, err := iso8601.ParseISO8601(want)
		if err != nil {
			t.Fatal(err)
		}
		got := sut.String()
		if want != got {
			t.Fatalf("Want %s, got %s", want, got)
		}
	}
}

func TestCanStringifyFractionalSeconds(t *testing.T) {
	cases := []struct {
		from string
		want string
	}{
		{"PT33.3444S", "PT33.3444S"},
		{"PT0.5S", "PT0.5S"},
		{"PT1.123S", "PT1.123S"},
		{"P343DT13H8M33.3444S", "P343DT13H8M33.3444S"},
		{"PT1.999999S", "PT1.999999S"},
		{"PT33.0S", "PT33S"}, // Should remove trailing zeros
		{"PT1.0S", "PT1S"},   // Should remove trailing zeros
	}
	for k, c := range cases {
		sut, err := iso8601.ParseISO8601(c.from)
		if err != nil {
			t.Fatalf("Case %d: failed to parse %s: %v", k, c.from, err)
		}
		got := sut.String()
		if c.want != got {
			t.Fatalf("Case %d: Want %s, got %s", k, c.want, got)
		}
	}
}

func TestCanMarshalJSON(t *testing.T) {
	s := "P1Y2M3W4DT5H6M7S"
	sut, _ := iso8601.ParseISO8601(s) //nolint:errcheck // Test helper, error handling tested elsewhere

	b, err := json.Marshal(sut)
	if err != nil {
		t.Fatal(err)
	}

	want := `"P1Y2M3W4DT5H6M7S"`
	got := string(b)
	if got != want {
		t.Fatalf("want=%s, got=%s", want, got)
	}
}

func TestCanUnmarshalJSON(t *testing.T) {
	j := []byte(`"P1Y2M3W4DT5H6M7S"`)
	var got iso8601.Duration
	err := json.Unmarshal(j, &got)
	if err != nil {
		t.Fatal(err)
	}

	s := "P1Y2M3W4DT5H6M7S"
	want, _ := iso8601.ParseISO8601(s) //nolint:errcheck // Test helper, error handling tested elsewhere

	if got != want {
		t.Fatalf("want=%+v, got=%+v", want, got)
	}
}

func TestCanRejectDurationInJSON(t *testing.T) {
	j := []byte(`"PZY"`)
	var got iso8601.Duration
	err := json.Unmarshal(j, &got)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}

func TestCanRejectBadJSON(t *testing.T) {
	j := []byte(`{"foo":"bar"}`)
	var got iso8601.Duration
	err := json.Unmarshal(j, &got)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}

func TestCanUnshift(t *testing.T) {
	cases := []struct {
		to       string
		duration iso8601.Duration
		want     string
	}{
		{"Jan 1, 2019 at 00:00:00", iso8601.Duration{Y: 1}, "Jan 1, 2018 at 00:00:00"},
		{"Feb 1, 2018 at 00:00:00", iso8601.Duration{M: 1}, "Jan 1, 2018 at 00:00:00"},
		{"Mar 1, 2018 at 00:00:00", iso8601.Duration{M: 2}, "Jan 1, 2018 at 00:00:00"},
		{"Jan 8, 2018 at 00:00:00", iso8601.Duration{W: 1}, "Jan 1, 2018 at 00:00:00"},
		{"Jan 2, 2018 at 00:00:00", iso8601.Duration{D: 1}, "Jan 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 01:00:00", iso8601.Duration{TH: 1}, "Jan 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:01:00", iso8601.Duration{TM: 1}, "Jan 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:01", iso8601.Duration{TS: 1}, "Jan 1, 2018 at 00:00:00"},
		{"Jun 9, 2028 at 05:10:06", iso8601.Duration{
			Y:  10,
			M:  5,
			D:  8,
			TH: 5,
			TM: 10,
			TS: 6,
		},
			"Jan 1, 2018 at 00:00:00",
		},
	}

	for k, c := range cases {
		to := makeTime(t, c.to)
		want := makeTime(t, c.want)

		got := c.duration.Unshift(to)
		if !want.Equal(got) {
			t.Fatalf("Case %d: want=%s, got=%s", k, want, got)
		}
	}
}

func TestCanUnshiftFractionalSeconds(t *testing.T) {
	// Test unshifting with fractional seconds
	// Using time.Unix for precise fractional second handling
	from := time.Unix(1514764801, 500000000) // Jan 1, 2018 00:00:01.5
	d := iso8601.Duration{TS: 0.5}
	got := d.Unshift(from)
	want := time.Unix(1514764801, 0) // Jan 1, 2018 00:00:01.0

	if !want.Equal(got) {
		t.Fatalf("want=%s, got=%s", want, got)
	}

	// Test unshifting 1.5 seconds
	from2 := time.Unix(1514764801, 0) // Jan 1, 2018 00:00:01.0
	d2 := iso8601.Duration{TS: 1.5}
	got2 := d2.Unshift(from2)
	want2 := time.Unix(1514764799, 500000000) // Dec 31, 2017 23:59:59.5

	if !want2.Equal(got2) {
		t.Fatalf("want=%s, got=%s", want2, got2)
	}
}

func TestCanParseNegativeDurations(t *testing.T) {
	cases := []struct {
		from string
		want iso8601.Duration
	}{
		{"-P1D", iso8601.Duration{D: -1}},
		{"-PT1H", iso8601.Duration{TH: -1}},
		{"-PT1M", iso8601.Duration{TM: -1}},
		{"-PT1S", iso8601.Duration{TS: -1}},
		{"-PT33.3444S", iso8601.Duration{TS: -33.3444}},
		{"-P1Y2M3DT4H5M6S", iso8601.Duration{Y: -1, M: -2, D: -3, TH: -4, TM: -5, TS: -6}},
	}

	for k, c := range cases {
		got, err := iso8601.ParseISO8601(c.from)
		if err != nil {
			t.Fatalf("Case %d: failed to parse %s: %v", k, c.from, err)
		}
		if !got.Equal(c.want) {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestCanStringifyNegativeDurations(t *testing.T) {
	cases := []struct {
		from string
		want string
	}{
		{"-P1D", "-P1D"},
		{"-PT1H", "-PT1H"},
		{"-PT1M", "-PT1M"},
		{"-PT1S", "-PT1S"},
		{"-PT33.3444S", "-PT33.3444S"},
		{"-P1Y2M3DT4H5M6S", "-P1Y2M3DT4H5M6S"},
	}

	for k, c := range cases {
		sut, err := iso8601.ParseISO8601(c.from)
		if err != nil {
			t.Fatalf("Case %d: failed to parse %s: %v", k, c.from, err)
		}
		got := sut.String()
		if c.want != got {
			t.Fatalf("Case %d: Want %s, got %s", k, c.want, got)
		}
	}
}

func TestCanAddDurations(t *testing.T) {
	cases := []struct {
		d1   iso8601.Duration
		d2   iso8601.Duration
		want iso8601.Duration
	}{
		{iso8601.Duration{D: 1}, iso8601.Duration{D: 2}, iso8601.Duration{D: 3}},
		{iso8601.Duration{TH: 1, TM: 30}, iso8601.Duration{TH: 2, TM: 15}, iso8601.Duration{TH: 3, TM: 45}},
		{iso8601.Duration{TS: 1.5}, iso8601.Duration{TS: 2.5}, iso8601.Duration{TS: 4.0}},
		{iso8601.Duration{D: 1}, iso8601.Duration{D: -1}, iso8601.Duration{D: 0}},
		{iso8601.Duration{Y: 1, M: 2}, iso8601.Duration{Y: 2, M: 3}, iso8601.Duration{Y: 3, M: 5}},
	}

	for k, c := range cases {
		got := c.d1.Add(c.d2)
		if !got.Equal(c.want) {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestCanSubtractDurations(t *testing.T) {
	cases := []struct {
		d1   iso8601.Duration
		d2   iso8601.Duration
		want iso8601.Duration
	}{
		{iso8601.Duration{D: 3}, iso8601.Duration{D: 2}, iso8601.Duration{D: 1}},
		{iso8601.Duration{TH: 3, TM: 45}, iso8601.Duration{TH: 2, TM: 15}, iso8601.Duration{TH: 1, TM: 30}},
		{iso8601.Duration{TS: 4.0}, iso8601.Duration{TS: 2.5}, iso8601.Duration{TS: 1.5}},
		{iso8601.Duration{D: 1}, iso8601.Duration{D: 2}, iso8601.Duration{D: -1}},
		{iso8601.Duration{Y: 3, M: 5}, iso8601.Duration{Y: 2, M: 3}, iso8601.Duration{Y: 1, M: 2}},
	}

	for k, c := range cases {
		got := c.d1.Subtract(c.d2)
		if !got.Equal(c.want) {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestCanMultiplyDurations(t *testing.T) {
	cases := []struct {
		d    iso8601.Duration
		n    int
		want iso8601.Duration
	}{
		{iso8601.Duration{D: 2}, 3, iso8601.Duration{D: 6}},
		{iso8601.Duration{TH: 1, TM: 30}, 2, iso8601.Duration{TH: 2, TM: 60}},
		{iso8601.Duration{TS: 1.5}, 2, iso8601.Duration{TS: 3.0}},
		{iso8601.Duration{D: 2}, -1, iso8601.Duration{D: -2}},
		{iso8601.Duration{Y: 1, M: 2}, 0, iso8601.Duration{}},
	}

	for k, c := range cases {
		got := c.d.Multiply(c.n)
		if !got.Equal(c.want) {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestCanCompareDurations(t *testing.T) {
	// Test Equal
	d1 := iso8601.Duration{D: 1, TH: 2, TM: 30, TS: 45.5}
	d2 := iso8601.Duration{D: 1, TH: 2, TM: 30, TS: 45.5}
	d3 := iso8601.Duration{D: 2, TH: 2, TM: 30, TS: 45.5}

	if !d1.Equal(d2) {
		t.Fatalf("Expected d1 and d2 to be equal")
	}
	if d1.Equal(d3) {
		t.Fatalf("Expected d1 and d3 to not be equal")
	}

	// Test LessThan (time-only)
	t1 := iso8601.Duration{TH: 1, TM: 30}
	t2 := iso8601.Duration{TH: 2, TM: 15}
	if !t1.LessThan(t2) {
		t.Fatalf("Expected t1 < t2")
	}
	if t2.LessThan(t1) {
		t.Fatalf("Expected t2 > t1")
	}

	// Test GreaterThan
	if !t2.GreaterThan(t1) {
		t.Fatalf("Expected t2 > t1")
	}
	if t1.GreaterThan(t2) {
		t.Fatalf("Expected t1 < t2")
	}

	// Test with fractional seconds
	ts1 := iso8601.Duration{TS: 1.5}
	ts2 := iso8601.Duration{TS: 2.5}
	if !ts1.LessThan(ts2) {
		t.Fatalf("Expected ts1 < ts2")
	}
}

func TestCanConvertToTimeDuration(t *testing.T) {
	cases := []struct {
		d    iso8601.Duration
		want time.Duration
	}{
		{iso8601.Duration{TH: 1}, time.Hour},
		{iso8601.Duration{TM: 1}, time.Minute},
		{iso8601.Duration{TS: 1}, time.Second},
		{iso8601.Duration{TS: 1.5}, 1500 * time.Millisecond},
		{iso8601.Duration{TH: 1, TM: 30, TS: 45}, 1*time.Hour + 30*time.Minute + 45*time.Second},
		{iso8601.Duration{TH: -1}, -time.Hour},
	}

	for k, c := range cases {
		got := c.d.ToTimeDuration()
		if got != c.want {
			t.Fatalf("Case %d: want=%v, got=%v", k, c.want, got)
		}
	}
}

func TestCanConvertFromTimeDuration(t *testing.T) {
	cases := []struct {
		td   time.Duration
		want iso8601.Duration
	}{
		{time.Hour, iso8601.Duration{TH: 1}},
		{time.Minute, iso8601.Duration{TM: 1}},
		{time.Second, iso8601.Duration{TS: 1}},
		{1500 * time.Millisecond, iso8601.Duration{TS: 1.5}},
		{1*time.Hour + 30*time.Minute + 45*time.Second, iso8601.Duration{TH: 1, TM: 30, TS: 45}},
		{-time.Hour, iso8601.Duration{TH: -1}},
		{90 * time.Second, iso8601.Duration{TM: 1, TS: 30}},
	}

	for k, c := range cases {
		got := iso8601.FromTimeDuration(c.td)
		if !got.Equal(c.want) {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestNegate(t *testing.T) {
	cases := []struct {
		d    iso8601.Duration
		want iso8601.Duration
	}{
		{iso8601.Duration{D: 1}, iso8601.Duration{D: -1}},
		{iso8601.Duration{D: -1}, iso8601.Duration{D: 1}},
		{iso8601.Duration{TH: 1, TM: 30, TS: 45.5}, iso8601.Duration{TH: -1, TM: -30, TS: -45.5}},
		{iso8601.Duration{}, iso8601.Duration{}},
	}

	for k, c := range cases {
		got := c.d.Negate()
		if !got.Equal(c.want) {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestIsNegative(t *testing.T) {
	cases := []struct {
		d    iso8601.Duration
		want bool
	}{
		{iso8601.Duration{D: -1}, true},
		{iso8601.Duration{D: 1}, false},
		{iso8601.Duration{TH: -1}, true},
		{iso8601.Duration{TS: -0.5}, true},
		{iso8601.Duration{}, false},
		{iso8601.Duration{D: 1, TH: -1}, true},
	}

	for k, c := range cases {
		got := c.d.IsNegative()
		if got != c.want {
			t.Fatalf("Case %d: want=%v, got=%v", k, c.want, got)
		}
	}
}
