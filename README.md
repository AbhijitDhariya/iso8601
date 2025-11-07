# ISO8601 Duration [![Go Report Card](https://goreportcard.com/badge/github.com/AbhijitDhariya/iso8601)](https://goreportcard.com/report/github.com/AbhijitDhariya/iso8601) [![GoDoc](https://godoc.org/github.com/AbhijitDhariya/iso8601?status.svg)](https://godoc.org/github.com/AbhijitDhariya/iso8601)

A Go package for parsing and manipulating ISO8601-formatted durations with support for fractional seconds and date/time shifting.

## Why This Library?

This library provides a comprehensive, well-tested solution for ISO8601 duration handling in Go with several key advantages:

- **✅ High Test Coverage** - 88.8% code coverage with comprehensive test suite
- **✅ ISO8601 Compliant** - Follows the ISO8601 standard for duration representation
- **✅ DST-Aware** - Correctly handles daylight saving time transitions when shifting dates
- **✅ Production Ready** - Clean API, comprehensive documentation, and robust error handling
- **✅ Feature Rich** - Supports fractional seconds, negative durations, arithmetic operations, and more
- **✅ Interoperable** - Seamless conversion to/from Go's `time.Duration`
- **✅ JSON Ready** - Built-in JSON marshaling/unmarshaling support
- **✅ Zero Dependencies** - Uses only Go standard library

## Features

- Parse ISO8601 duration strings (e.g., `P1Y2M3DT4H5M6.5S`)
- Support for fractional seconds (e.g., `P343DT13H8M33.3444S`)
- Support for negative durations (e.g., `-P1D`, `-PT1H`)
- Duration arithmetic (Add, Subtract, Multiply)
- Comparison methods (Equal, LessThan, GreaterThan)
- Conversion to/from Go's `time.Duration`
- Shift dates/times forward and backward
- JSON marshaling/unmarshaling
- Handles DST transitions correctly

## Basic Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/AbhijitDhariya/iso8601"
)

func main() {
	d, _ := iso8601.ParseISO8601("P1D")
	today := time.Now()
	tomorrow := d.Shift(today)
	yesterday := d.Unshift(today)
	fmt.Println(today.Format("Jan _2"))
	fmt.Println(tomorrow.Format("Jan _2"))
	fmt.Println(yesterday.Format("Jan _2"))
}
```

## Fractional Seconds

The package supports fractional seconds in ISO8601 duration strings:

```go
d, _ := iso8601.ParseISO8601("P343DT13H8M33.3444S")
fmt.Println(d.String()) // Output: P343DT13H8M33.3444S

d2, _ := iso8601.ParseISO8601("PT0.5S")
fmt.Println(d2.String()) // Output: PT0.5S
```

## Negative Durations

The package supports negative durations with a leading minus sign:

```go
d, _ := iso8601.ParseISO8601("-P1D")
fmt.Println(d.String()) // Output: -P1D

d2, _ := iso8601.ParseISO8601("-PT1H30M")
fmt.Println(d2.String()) // Output: -PT1H30M

// Check if a duration is negative
if d.IsNegative() {
    fmt.Println("Duration is negative")
}

// Negate a duration
positive := d.Negate()
fmt.Println(positive.String()) // Output: P1D
```

### Use Cases for Negative Durations

Negative durations are particularly useful in several scenarios:

#### 1. **Time Differences and Deltas**

When calculating the difference between two times, negative durations naturally represent when the end time is before the start time:

```go
start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
end := time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC) // One day earlier

// Calculate duration between times
diff := end.Sub(start) // This will be negative
d := iso8601.FromTimeDuration(diff)
fmt.Println(d.String()) // Output: -PT24H
```

#### 2. **Deadline Tracking**

Represent time until or past a deadline:

```go
deadline := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
now := time.Now()

// Time until deadline (could be negative if past due)
remaining := deadline.Sub(now)
d := iso8601.FromTimeDuration(remaining)

if d.IsNegative() {
    fmt.Printf("Deadline passed %s ago\n", d.Negate().String())
} else {
    fmt.Printf("Time remaining: %s\n", d.String())
}
```

#### 3. **Offset Calculations**

Represent timezone offsets or scheduling offsets that go backward:

```go
// Represent a timezone offset (e.g., UTC-5)
offset, _ := iso8601.ParseISO8601("-PT5H")
baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
adjustedTime := offset.Shift(baseTime) // Goes backward 5 hours
```

#### 4. **API Responses and Data Models**

Many APIs return durations that can be negative (e.g., "time until event" or "time since event"):

```go
// Example: API returns "-PT2H" meaning "2 hours ago"
apiResponse := "-PT2H"
d, _ := iso8601.ParseISO8601(apiResponse)

if d.IsNegative() {
    eventTime := d.Shift(time.Now()) // Calculate when event occurred
    fmt.Printf("Event occurred at: %s\n", eventTime.Format(time.RFC3339))
}
```

#### 5. **Arithmetic with Time Differences**

When subtracting durations, negative results are natural:

```go
d1, _ := iso8601.ParseISO8601("PT1H")
d2, _ := iso8601.ParseISO8601("PT2H")

result := d1.Subtract(d2) // PT1H - PT2H = -PT1H
fmt.Println(result.String()) // Output: -PT1H
```

#### 6. **Scientific and Measurement Contexts**

In contexts where negative values represent going backward or reversing direction:

```go
// Example: Representing a time shift that goes backward
backwardShift, _ := iso8601.ParseISO8601("-P1DT6H")
originalTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
shiftedTime := backwardShift.Shift(originalTime)
// Result: 1 day and 6 hours earlier
```

## Duration Arithmetic

Perform arithmetic operations on durations:

```go
d1, _ := iso8601.ParseISO8601("P1DT2H")
d2, _ := iso8601.ParseISO8601("P2DT3H")

// Add durations
sum := d1.Add(d2)
fmt.Println(sum.String()) // Output: P3DT5H

// Subtract durations
diff := d2.Subtract(d1)
fmt.Println(diff.String()) // Output: P1DT1H

// Multiply a duration
doubled := d1.Multiply(2)
fmt.Println(doubled.String()) // Output: P2DT4H
```

**Note:** Arithmetic operations perform component-wise addition/subtraction. For durations with months/years, the result may not represent the exact calendar duration due to variable month lengths.

## Comparison Methods

Compare durations using the comparison methods:

```go
d1, _ := iso8601.ParseISO8601("PT1H30M")
d2, _ := iso8601.ParseISO8601("PT2H")

// Check equality
if d1.Equal(d2) {
    fmt.Println("Durations are equal")
}

// Compare durations
if d1.LessThan(d2) {
    fmt.Println("d1 is less than d2")
}

if d2.GreaterThan(d1) {
    fmt.Println("d2 is greater than d1")
}
```

**Note:** Comparison is most meaningful for time-only durations (no years/months/weeks/days). For durations with date components, the result may be ambiguous due to variable month lengths.

## Conversion to/from time.Duration

Convert between ISO8601 durations and Go's `time.Duration`:

```go
// Convert ISO8601 duration to time.Duration (time component only)
d, _ := iso8601.ParseISO8601("PT1H30M45.5S")
td := d.ToTimeDuration()
fmt.Println(td) // Output: 1h30m45.5s

// Create ISO8601 duration from time.Duration
td2 := 2*time.Hour + 30*time.Minute
d2 := iso8601.FromTimeDuration(td2)
fmt.Println(d2.String()) // Output: PT2H30M

// Works with fractional seconds
td3 := 1500 * time.Millisecond
d3 := iso8601.FromTimeDuration(td3)
fmt.Println(d3.String()) // Output: PT1.5S
```

**Note:** `ToTimeDuration()` only converts the time component (hours, minutes, seconds). Date components (years, months, weeks, days) are ignored. `FromTimeDuration()` only sets the time component; date components are zero.

## Why Does This Package Exist?

> Why can't we just use a `time.Duration` and `time.Add`?

A very reasonable question.

The code below repeatedly adds 24 hours to a `time.Time`. You might expect the time on that date to stay the same, but [_there are not always 24 hours in a day_](http://infiniteundo.com/post/25326999628/falsehoods-programmers-believe-about-time). When the clocks change in New York, the time will skew by an hour. As you can see from the output, `iso8601.Duration.Shift()` can increment the date without shifting the time.

```go
package main

import (
	"fmt"
	"time"

	"github.com/AbhijitDhariya/iso8601"
)

func main() {
	loc, _ := time.LoadLocation("America/New_York")
	d, _ := iso8601.ParseISO8601("P1D")
	t1, _ := time.ParseInLocation("Jan 2, 2006 at 3:04pm", "Jan 1, 2006 at 3:04pm", loc)
	t2 := t1
	for i := 0; i < 365; i++ {
		t1 = t1.Add(24 * time.Hour)
		t2 = d.Shift(t2)
		fmt.Printf("time.Add:%d    Duration.Shift:%d\n", t1.Hour(), t2.Hour())
	}
}

// Outputs
// time.Add:15    Duration.Shift:15
// time.Add:15    Duration.Shift:15
// time.Add:15    Duration.Shift:15
// ...
// time.Add:16    Duration.Shift:15
// time.Add:16    Duration.Shift:15
// time.Add:16    Duration.Shift:15
// ...
```

## Months and Date Rolling

Months are tricky. Shifting by months uses `time.AddDate()`, which is great. However, be aware of how differing days in the month are accommodated. Dates will 'roll over' if the month you're shifting to has fewer days. e.g. if you start on Jan 30th and repeat every "P1M", you'll get this:

```
Jan 30, 2006
Mar 2, 2006
Apr 2, 2006
May 2, 2006
Jun 2, 2006
Jul 2, 2006
Aug 2, 2006
Sep 2, 2006
Oct 2, 2006
Nov 2, 2006
Dec 2, 2006
Jan 2, 2007
```

## API

### ParseISO8601

Parses an ISO8601 duration string:

```go
d, err := iso8601.ParseISO8601("P1Y2M3DT4H5M6.5S")
```

### Shift

Returns a `time.Time`, shifted forward by the duration from the given start:

```go
d, _ := iso8601.ParseISO8601("P1D")
tomorrow := d.Shift(time.Now())
```

**Note:** Shift uses `time.AddDate` for years, months, weeks, and days, and so shares its limitations. In particular, shifting by months is not recommended unless the start date is before the 28th of the month. Otherwise, dates will roll over, e.g. Aug 31 + P1M = Oct 1.

Week and Day values will be combined as W*7 + D.

### Unshift

Returns a `time.Time`, shifted back by the duration from the given start:

```go
d, _ := iso8601.ParseISO8601("P1D")
yesterday := d.Unshift(time.Now())
```

**Note:** Unshift uses `time.AddDate` for years, months, weeks, and days, and so shares its limitations. In particular, shifting back by months is not recommended unless the start date is before the 28th of the month. Otherwise, dates will roll over, e.g. Oct 1 - P1M = Aug 31.

Week and Day values will be combined as W*7 + D.

### String

Returns an ISO8601 representation of the duration:

```go
d, _ := iso8601.ParseISO8601("P1Y2M3DT4H5M6.5S")
fmt.Println(d.String()) // Output: P1Y2M3DT4H5M6.5S
```

### Arithmetic Operations

Perform arithmetic on durations:

```go
d1, _ := iso8601.ParseISO8601("P1DT2H")
d2, _ := iso8601.ParseISO8601("P2DT3H")

sum := d1.Add(d2)           // Add durations
diff := d2.Subtract(d1)      // Subtract durations
doubled := d1.Multiply(2)    // Multiply by scalar
```

### Comparison Methods

Compare durations:

```go
d1, _ := iso8601.ParseISO8601("PT1H")
d2, _ := iso8601.ParseISO8601("PT2H")

if d1.Equal(d2) { ... }        // Check equality
if d1.LessThan(d2) { ... }    // Less than
if d2.GreaterThan(d1) { ... }  // Greater than
```

### Conversion Methods

Convert to/from Go's `time.Duration`:

```go
d, _ := iso8601.ParseISO8601("PT1H30M")
td := d.ToTimeDuration()                    // Convert to time.Duration
d2 := iso8601.FromTimeDuration(time.Hour)  // Create from time.Duration
```

### Helper Methods

```go
d, _ := iso8601.ParseISO8601("-P1D")

if d.IsNegative() { ... }  // Check if negative
negated := d.Negate()      // Negate duration
```

### JSON Support

The `Duration` type implements `json.Marshaler` and `json.Unmarshaler`:

```go
d, _ := iso8601.ParseISO8601("P1D")
data, _ := json.Marshal(d)
// data is []byte(`"P1D"`)

var d2 iso8601.Duration
json.Unmarshal(data, &d2)
```

## Key Strengths

### 1. **Robust Date/Time Handling**

Unlike simple `time.Duration` arithmetic, this library correctly handles:
- **Daylight Saving Time (DST) transitions** - Shifting by days maintains the time of day even when clocks change
- **Variable month lengths** - Properly handles months with different numbers of days
- **Leap years** - Correctly accounts for leap years when shifting dates

### 2. **Comprehensive Feature Set**

- **Fractional Seconds** - Full support for sub-second precision (e.g., `PT33.3444S`)
- **Negative Durations** - Handle backward time shifts and time differences naturally
- **Arithmetic Operations** - Add, subtract, and multiply durations
- **Comparison Methods** - Compare durations for equality and ordering
- **Conversion Utilities** - Easy conversion to/from Go's `time.Duration`

### 3. **Clean and Intuitive API**

The API is designed to be simple and intuitive:

```go
// Parse and use - straightforward
d, _ := iso8601.ParseISO8601("P1D")
tomorrow := d.Shift(time.Now())

// Arithmetic - natural and readable
sum := d1.Add(d2)
diff := d2.Subtract(d1)

// Conversion - seamless
td := d.ToTimeDuration()
d2 := iso8601.FromTimeDuration(time.Hour)
```

### 4. **Well-Tested and Documented**

- **88.8% test coverage** with comprehensive test cases
- **Extensive documentation** with practical examples
- **Edge case handling** for DST, negative values, fractional seconds
- **Clear documentation** of limitations and best practices

### 5. **Production Ready**

- **Zero external dependencies** - Uses only Go standard library
- **JSON support** - Built-in marshaling/unmarshaling
- **Error handling** - Proper error returns for invalid input
- **Performance** - Efficient regex-based parsing

### 6. **Standards Compliant**

- Follows ISO8601 duration format specification
- Handles all standard components (years, months, weeks, days, hours, minutes, seconds)
- Supports optional time separator (`T`)
- Proper formatting with trailing zero removal

## Installation

```bash
go get github.com/AbhijitDhariya/iso8601
```

## License

See LICENSE file.
