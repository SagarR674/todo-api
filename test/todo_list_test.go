package test

import (
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type pagedTodos struct {
	Items      []todoView `json:"items"`
	Pagination struct {
		Page       int   `json:"page"`
		Limit      int   `json:"limit"`
		Total      int64 `json:"total"`
		TotalPages int   `json:"total_pages"`
		HasNext    bool  `json:"has_next"`
		HasPrev    bool  `json:"has_prev"`
	} `json:"pagination"`
}

// seedTodos creates a known set of todos for one user and returns their token.
func seedListData(t *testing.T, c *apiClient) string {
	t.Helper()
	token := c.authUser("list@example.com")

	todos := []fiber.Map{
		{"title": "backend api work", "status": "pending", "priority": "high", "due_date": "2026-03-01"},
		{"title": "frontend polish", "status": "in_progress", "priority": "medium", "due_date": "2026-01-15"},
		{"title": "write backend tests", "status": "pending", "priority": "low", "due_date": "2026-06-20"},
		{"title": "deploy release", "status": "completed", "priority": "high", "due_date": "2026-02-10"},
		{"title": "read documentation", "status": "pending", "priority": "medium"},
	}
	for _, body := range todos {
		if res := c.createTodo(token, body); res.status != fiber.StatusCreated {
			t.Fatalf("seed todo: %d (%s)", res.status, res.raw)
		}
	}
	return token
}

func list(t *testing.T, c *apiClient, token, query string) pagedTodos {
	t.Helper()
	res := c.do(fiber.MethodGet, "/api/todos?"+query, token, nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("list ?%s: status %d (%s)", query, res.status, res.raw)
	}
	var page pagedTodos
	res.decodeData(t, &page)
	return page
}

func TestList_Pagination(t *testing.T) {
	c := newClient(t)
	token := seedListData(t, c)

	p1 := list(t, c, token, "page=1&limit=2")
	if p1.Pagination.Total != 5 || p1.Pagination.TotalPages != 3 {
		t.Errorf("meta = %+v", p1.Pagination)
	}
	if len(p1.Items) != 2 {
		t.Errorf("page 1 items = %d, want 2", len(p1.Items))
	}
	if !p1.Pagination.HasNext || p1.Pagination.HasPrev {
		t.Errorf("page 1 next/prev = %v/%v", p1.Pagination.HasNext, p1.Pagination.HasPrev)
	}

	p3 := list(t, c, token, "page=3&limit=2")
	if len(p3.Items) != 1 {
		t.Errorf("page 3 items = %d, want 1", len(p3.Items))
	}
	if p3.Pagination.HasNext || !p3.Pagination.HasPrev {
		t.Errorf("page 3 next/prev = %v/%v", p3.Pagination.HasNext, p3.Pagination.HasPrev)
	}

	// out-of-range page -> empty items, still 200
	p9 := list(t, c, token, "page=9&limit=2")
	if len(p9.Items) != 0 {
		t.Errorf("page 9 items = %d, want 0", len(p9.Items))
	}

	// limit clamps to 100
	big := list(t, c, token, "limit=99999")
	if big.Pagination.Limit != 100 {
		t.Errorf("limit clamp = %d, want 100", big.Pagination.Limit)
	}

	// bad params fall back to defaults
	bad := list(t, c, token, "page=-1&limit=abc")
	if bad.Pagination.Page != 1 || bad.Pagination.Limit != 10 {
		t.Errorf("bad params meta = %+v", bad.Pagination)
	}
}

func TestList_FilterByStatus(t *testing.T) {
	c := newClient(t)
	token := seedListData(t, c)

	p := list(t, c, token, "status=pending&limit=100")
	if p.Pagination.Total != 3 {
		t.Errorf("pending total = %d, want 3", p.Pagination.Total)
	}
	for _, it := range p.Items {
		if it.Status != "pending" {
			t.Errorf("got status %q in pending filter", it.Status)
		}
	}

	// invalid status value -> 400
	res := c.do(fiber.MethodGet, "/api/todos?status=bogus", token, nil)
	if res.status != fiber.StatusBadRequest {
		t.Errorf("invalid status filter = %d, want 400", res.status)
	}
}

func TestList_FilterByPriority(t *testing.T) {
	c := newClient(t)
	token := seedListData(t, c)

	p := list(t, c, token, "priority=high&limit=100")
	if p.Pagination.Total != 2 {
		t.Errorf("high total = %d, want 2", p.Pagination.Total)
	}

	res := c.do(fiber.MethodGet, "/api/todos?priority=bogus", token, nil)
	if res.status != fiber.StatusBadRequest {
		t.Errorf("invalid priority filter = %d, want 400", res.status)
	}
}

func TestList_SearchByTitle(t *testing.T) {
	c := newClient(t)
	token := seedListData(t, c)

	p := list(t, c, token, "search=backend&limit=100")
	if p.Pagination.Total != 2 {
		t.Errorf("search 'backend' total = %d, want 2 (%+v)", p.Pagination.Total, titles(p.Items))
	}

	none := list(t, c, token, "search=nonexistentxyz")
	if none.Pagination.Total != 0 {
		t.Errorf("search miss total = %d, want 0", none.Pagination.Total)
	}
}

func TestList_CombinedFilters(t *testing.T) {
	c := newClient(t)
	token := seedListData(t, c)

	p := list(t, c, token, "status=pending&priority=high&limit=100")
	if p.Pagination.Total != 1 {
		t.Fatalf("pending+high total = %d, want 1 (%+v)", p.Pagination.Total, titles(p.Items))
	}
	if p.Items[0].Title != "backend api work" {
		t.Errorf("got %q", p.Items[0].Title)
	}
}

func TestList_Sorting(t *testing.T) {
	c := newClient(t)
	token := seedListData(t, c)

	// due_date ascending: todos without a due date sort first (NULL), then chronological
	asc := list(t, c, token, "sort=due_date&order=asc&limit=100")
	var lastDue string
	for _, it := range asc.Items {
		if it.DueDate == "" {
			continue
		}
		if lastDue != "" && it.DueDate < lastDue {
			t.Errorf("due_date not ascending: %s after %s", it.DueDate, lastDue)
		}
		lastDue = it.DueDate
	}

	desc := list(t, c, token, "sort=due_date&order=desc&limit=100")
	if len(desc.Items) > 0 && len(asc.Items) > 0 {
		if desc.Items[0].DueDate != "2026-06-20" {
			t.Errorf("desc first due_date = %q, want 2026-06-20", desc.Items[0].DueDate)
		}
	}

	// priority ascending uses the MySQL ENUM order low < medium < high
	pr := list(t, c, token, "sort=priority&order=asc&limit=100")
	rank := map[string]int{"low": 0, "medium": 1, "high": 2}
	last := -1
	for _, it := range pr.Items {
		if rank[it.Priority] < last {
			t.Errorf("priority not ascending: %q after rank %d", it.Priority, last)
		}
		last = rank[it.Priority]
	}

	// an unknown / injection-y sort field must be ignored, not error
	res := c.do(fiber.MethodGet, "/api/todos?"+url.Values{"sort": {"id; DROP TABLE todos"}}.Encode(), token, nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("unknown sort field status = %d, want 200 (%s)", res.status, res.raw)
	}
	// table still there
	still := list(t, c, token, "limit=100")
	if still.Pagination.Total != 5 {
		t.Errorf("todos table affected by sort injection: total = %d", still.Pagination.Total)
	}
}

func titles(items []todoView) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Title
	}
	return out
}
