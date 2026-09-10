package models

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	d, err := ParseDate("2026-09-15")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if d.Year() != 2026 || d.Month() != time.September || d.Day() != 15 {
		t.Errorf("got %v", d.Time)
	}
}

func TestParseDate_Invalid(t *testing.T) {
	for _, s := range []string{"15-09-2026", "2026/09/15", "not-a-date", "2026-13-40", ""} {
		if _, err := ParseDate(s); err == nil {
			t.Errorf("ParseDate(%q) should fail", s)
		} else if !errors.Is(err, ErrInvalidDate) {
			t.Errorf("ParseDate(%q) error should wrap ErrInvalidDate, got %v", s, err)
		}
	}
}

func TestDate_JSONRoundTrip(t *testing.T) {
	type wrap struct {
		Due *Date `json:"due"`
	}

	var w wrap
	if err := json.Unmarshal([]byte(`{"due":"2026-01-02"}`), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if w.Due == nil || w.Due.String() != "2026-01-02" {
		t.Fatalf("got %+v", w.Due)
	}

	out, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `{"due":"2026-01-02"}` {
		t.Errorf("marshal = %s", out)
	}
}

func TestDate_JSONNull(t *testing.T) {
	var d Date
	b, _ := d.MarshalJSON()
	if string(b) != "null" {
		t.Errorf("zero date should marshal to null, got %s", b)
	}

	if err := json.Unmarshal([]byte("null"), &d); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if !d.Time.IsZero() {
		t.Error("null should unmarshal to the zero value")
	}
}

func TestDate_ValueAndScan(t *testing.T) {
	d, _ := ParseDate("2026-06-10")

	v, err := d.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != "2026-06-10" {
		t.Errorf("Value = %v", v)
	}

	var zero Date
	if zv, _ := zero.Value(); zv != nil {
		t.Errorf("zero Value should be nil, got %v", zv)
	}

	tests := []any{
		time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		[]byte("2026-06-10"),
		"2026-06-10",
	}
	for _, in := range tests {
		var got Date
		if err := got.Scan(in); err != nil {
			t.Fatalf("Scan(%T): %v", in, err)
		}
		if got.String() != "2026-06-10" {
			t.Errorf("Scan(%T) => %s", in, got.String())
		}
	}

	var nilDate Date
	if err := nilDate.Scan(nil); err != nil || !nilDate.Time.IsZero() {
		t.Errorf("Scan(nil) => %v, %v", nilDate, err)
	}

	var bad Date
	if err := bad.Scan(12345); err == nil {
		t.Error("Scan(int) should fail")
	}
}
