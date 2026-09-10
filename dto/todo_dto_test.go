package dto

import "testing"

func TestListTodosQuery_Normalize_Defaults(t *testing.T) {
	q := &ListTodosQuery{}
	q.Normalize()

	if q.Page != 1 {
		t.Errorf("Page = %d, want 1", q.Page)
	}
	if q.Limit != 10 {
		t.Errorf("Limit = %d, want 10", q.Limit)
	}
	if q.Sort != "created_at" {
		t.Errorf("Sort = %q, want created_at", q.Sort)
	}
	if q.Order != "desc" {
		t.Errorf("Order = %q, want desc", q.Order)
	}
}

func TestListTodosQuery_Normalize_Clamps(t *testing.T) {
	q := &ListTodosQuery{Page: -5, Limit: 9999}
	q.Normalize()
	if q.Page != 1 {
		t.Errorf("negative page should clamp to 1, got %d", q.Page)
	}
	if q.Limit != 100 {
		t.Errorf("limit should clamp to 100, got %d", q.Limit)
	}

	q = &ListTodosQuery{Limit: 0}
	q.Normalize()
	if q.Limit != 10 {
		t.Errorf("zero limit should default to 10, got %d", q.Limit)
	}
}

func TestListTodosQuery_Normalize_SortWhitelist(t *testing.T) {
	q := &ListTodosQuery{Sort: "password); DROP TABLE todos;--", Order: "ASC"}
	q.Normalize()
	if q.Sort != "created_at" {
		t.Errorf("unknown sort field should fall back to created_at, got %q", q.Sort)
	}
	if q.Order != "asc" {
		t.Errorf("Order = %q, want asc", q.Order)
	}
}

func TestListTodosQuery_OrderClause(t *testing.T) {
	q := &ListTodosQuery{Sort: "due_date", Order: "asc"}
	q.Normalize()
	if got := q.OrderClause(); got != "todos.due_date asc" {
		t.Errorf("OrderClause = %q", got)
	}
}

func TestListTodosQuery_Offset(t *testing.T) {
	cases := []struct{ page, limit, want int }{
		{1, 10, 0},
		{2, 10, 10},
		{5, 25, 100},
	}
	for _, c := range cases {
		q := &ListTodosQuery{Page: c.page, Limit: c.limit}
		if got := q.Offset(); got != c.want {
			t.Errorf("Offset(page=%d,limit=%d) = %d, want %d", c.page, c.limit, got, c.want)
		}
	}
}

func TestNewPagination(t *testing.T) {
	cases := []struct {
		page, limit        int
		total              int64
		wantPages          int
		wantNext, wantPrev bool
	}{
		{1, 10, 0, 0, false, false},
		{1, 10, 5, 1, false, false},
		{1, 10, 25, 3, true, false},
		{2, 10, 25, 3, true, true},
		{3, 10, 25, 3, false, true},
		{1, 10, 10, 1, false, false},
	}
	for _, c := range cases {
		p := NewPagination(c.page, c.limit, c.total)
		if p.TotalPages != c.wantPages {
			t.Errorf("total=%d: TotalPages = %d, want %d", c.total, p.TotalPages, c.wantPages)
		}
		if p.HasNext != c.wantNext {
			t.Errorf("page=%d total=%d: HasNext = %v, want %v", c.page, c.total, p.HasNext, c.wantNext)
		}
		if p.HasPrev != c.wantPrev {
			t.Errorf("page=%d total=%d: HasPrev = %v, want %v", c.page, c.total, p.HasPrev, c.wantPrev)
		}
	}
}
