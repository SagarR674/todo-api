package utils

import "testing"

type sample struct {
	Name     string `json:"name"     validate:"required,min=2"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Status   string `json:"status"   validate:"omitempty,oneof=pending in_progress completed"`
	DueDate  string `json:"due_date" validate:"omitempty,datetime=2006-01-02"`
	Hidden   string `json:"-"        validate:"required"`
}

func TestValidateStruct_Valid(t *testing.T) {
	in := sample{Name: "Al", Email: "al@example.com", Password: "longenough", Status: "pending", Hidden: "x"}
	if errs := ValidateStruct(in); errs != nil {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateStruct_CollectsFieldErrors(t *testing.T) {
	in := sample{Name: "A", Email: "nope", Password: "short", Status: "weird"}
	errs := ValidateStruct(in)
	if errs == nil {
		t.Fatal("expected validation errors")
	}

	for _, field := range []string{"name", "email", "password", "status"} {
		if _, ok := errs[field]; !ok {
			t.Errorf("missing error for %q (got %v)", field, errs)
		}
	}
}

func TestValidateStruct_UsesJSONTagNames(t *testing.T) {
	errs := ValidateStruct(sample{})
	if _, ok := errs["Name"]; ok {
		t.Error("errors should be keyed by json tag, not Go field name")
	}
	if _, ok := errs["name"]; !ok {
		t.Error("expected json tag key 'name'")
	}
}

func TestValidateStruct_Messages(t *testing.T) {
	errs := ValidateStruct(sample{Name: "A", Email: "x@y.z", Password: "abcdefgh", Status: "bad", Hidden: "h"})
	if got := errs["status"]; got != "must be one of: pending, in_progress, completed" {
		t.Errorf("status message = %q", got)
	}
	errs = ValidateStruct(sample{Email: "x@y.z", Password: "abcdefgh", Hidden: "h"})
	if got := errs["name"]; got != "name is required" {
		t.Errorf("name message = %q", got)
	}
}

func TestValidateStruct_DateMessage(t *testing.T) {
	errs := ValidateStruct(sample{Name: "Al", Email: "x@y.z", Password: "abcdefgh", DueDate: "15-09-2026", Hidden: "h"})
	if got := errs["due_date"]; got != "must be a valid date in YYYY-MM-DD format" {
		t.Errorf("due_date message = %q", got)
	}
}
