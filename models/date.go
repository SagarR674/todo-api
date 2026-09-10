package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DateLayout is the wire format for date-only values ("2026-09-15").
const DateLayout = "2006-01-02"

// ErrInvalidDate is returned when a date string is not in YYYY-MM-DD form.
var ErrInvalidDate = errors.New("invalid date, expected format YYYY-MM-DD")

// Date is a date-only value. It marshals to/from JSON as "YYYY-MM-DD" and is
// stored in a MySQL DATE column, keeping the API contract clean for fields such
// as a todo's due date.
type Date struct {
	time.Time
}

// NewDate wraps a time.Time as a Date (time component ignored).
func NewDate(t time.Time) Date {
	return Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

// ParseDate parses a "YYYY-MM-DD" string.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("%w: %q", ErrInvalidDate, s)
	}
	return Date{Time: t}, nil
}

func (d Date) String() string { return d.Format(DateLayout) }

// MarshalJSON renders the date as "YYYY-MM-DD", or null when zero.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Format(DateLayout) + `"`), nil
}

// UnmarshalJSON accepts "YYYY-MM-DD" or null.
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// Value implements driver.Valuer for database writes.
func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Format(DateLayout), nil
}

// Scan implements sql.Scanner for database reads.
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case []byte:
		return d.UnmarshalJSON([]byte(`"` + string(v) + `"`))
	case string:
		return d.UnmarshalJSON([]byte(`"` + v + `"`))
	default:
		return errors.New("unsupported Scan type for models.Date")
	}
}

// GormDataType tells GORM to use a DATE column.
func (Date) GormDataType() string { return "date" }
