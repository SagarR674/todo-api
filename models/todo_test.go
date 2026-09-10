package models

import "testing"

func TestStatus_IsValid(t *testing.T) {
	valid := []Status{StatusPending, StatusInProgress, StatusCompleted}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("%q should be valid", s)
		}
	}
	for _, s := range []Status{"", "done", "PENDING", "in progress"} {
		if s.IsValid() {
			t.Errorf("%q should be invalid", s)
		}
	}
}

func TestPriority_IsValid(t *testing.T) {
	for _, p := range []Priority{PriorityLow, PriorityMedium, PriorityHigh} {
		if !p.IsValid() {
			t.Errorf("%q should be valid", p)
		}
	}
	for _, p := range []Priority{"", "urgent", "HIGH", "normal"} {
		if p.IsValid() {
			t.Errorf("%q should be invalid", p)
		}
	}
}

func TestValidEnumSlices(t *testing.T) {
	if len(ValidStatuses) != 3 || len(ValidPriorities) != 3 {
		t.Fatalf("expected 3 statuses and 3 priorities, got %d / %d", len(ValidStatuses), len(ValidPriorities))
	}
}
