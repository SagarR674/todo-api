package test

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type todoView struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Priority   string `json:"priority"`
	DueDate    string `json:"due_date"`
	Categories []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"categories"`
}

func TestTodo_CreateWithDefaults(t *testing.T) {
	c := newClient(t)
	token := c.authUser("t1@example.com")

	res := c.createTodo(token, fiber.Map{"title": "Only a title"})
	if res.status != fiber.StatusCreated {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var todo todoView
	res.decodeData(t, &todo)
	if todo.Status != "pending" || todo.Priority != "medium" {
		t.Errorf("defaults wrong: %+v", todo)
	}
	if todo.ID == 0 {
		t.Error("expected an id")
	}
}

func TestTodo_CreateFull(t *testing.T) {
	c := newClient(t)
	token := c.authUser("t2@example.com")

	res := c.createTodo(token, fiber.Map{
		"title":       "Complete backend assignment",
		"description": "Develop Todo APIs using Golang and Fiber",
		"status":      "in_progress",
		"priority":    "high",
		"due_date":    "2026-09-15",
	})
	if res.status != fiber.StatusCreated {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var todo todoView
	res.decodeData(t, &todo)
	if todo.Status != "in_progress" || todo.Priority != "high" || todo.DueDate != "2026-09-15" {
		t.Errorf("unexpected todo %+v", todo)
	}
}

func TestTodo_CreateValidation(t *testing.T) {
	c := newClient(t)
	token := c.authUser("t3@example.com")

	cases := []struct {
		name  string
		body  fiber.Map
		field string
	}{
		{"empty title", fiber.Map{"title": ""}, "title"},
		{"missing title", fiber.Map{"description": "x"}, "title"},
		{"bad status", fiber.Map{"title": "x", "status": "done"}, "status"},
		{"bad priority", fiber.Map{"title": "x", "priority": "urgent"}, "priority"},
		{"bad due date", fiber.Map{"title": "x", "due_date": "15-09-2026"}, "due_date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := c.createTodo(token, tc.body)
			if res.status != fiber.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (%s)", res.status, res.raw)
			}
			if _, ok := res.body.Errors[tc.field]; !ok {
				t.Errorf("expected error for %q, got %v", tc.field, res.body.Errors)
			}
		})
	}
}

func TestTodo_GetByID(t *testing.T) {
	c := newClient(t)
	token := c.authUser("t4@example.com")
	created := c.createTodo(token, fiber.Map{"title": "find me"})
	var made todoView
	created.decodeData(t, &made)

	res := c.do(fiber.MethodGet, fmt.Sprintf("/api/todos/%d", made.ID), token, nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var got todoView
	res.decodeData(t, &got)
	if got.ID != made.ID || got.Title != "find me" {
		t.Errorf("got %+v", got)
	}

	// missing id
	res = c.do(fiber.MethodGet, "/api/todos/99999", token, nil)
	if res.status != fiber.StatusNotFound {
		t.Errorf("missing todo status = %d, want 404", res.status)
	}
	// non-numeric id
	res = c.do(fiber.MethodGet, "/api/todos/abc", token, nil)
	if res.status != fiber.StatusBadRequest {
		t.Errorf("non-numeric id status = %d, want 400", res.status)
	}
}

func TestTodo_Update(t *testing.T) {
	c := newClient(t)
	token := c.authUser("t5@example.com")
	created := c.createTodo(token, fiber.Map{"title": "before", "priority": "low"})
	var made todoView
	created.decodeData(t, &made)

	res := c.do(fiber.MethodPut, fmt.Sprintf("/api/todos/%d", made.ID), token, fiber.Map{
		"title": "after", "status": "completed",
	})
	if res.status != fiber.StatusOK {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var got todoView
	res.decodeData(t, &got)
	if got.Title != "after" || got.Status != "completed" {
		t.Errorf("update not applied: %+v", got)
	}
	if got.Priority != "low" {
		t.Errorf("priority should be unchanged, got %q", got.Priority)
	}
}

func TestTodo_UpdateStatus(t *testing.T) {
	c := newClient(t)
	token := c.authUser("t6@example.com")
	created := c.createTodo(token, fiber.Map{"title": "x"})
	var made todoView
	created.decodeData(t, &made)

	res := c.do(fiber.MethodPatch, fmt.Sprintf("/api/todos/%d/status", made.ID), token, fiber.Map{"status": "completed"})
	if res.status != fiber.StatusOK {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var got todoView
	res.decodeData(t, &got)
	if got.Status != "completed" {
		t.Errorf("status = %q", got.Status)
	}

	// invalid status value
	res = c.do(fiber.MethodPatch, fmt.Sprintf("/api/todos/%d/status", made.ID), token, fiber.Map{"status": "nope"})
	if res.status != fiber.StatusBadRequest {
		t.Errorf("bad status value = %d, want 400", res.status)
	}
}

func TestTodo_SoftDelete(t *testing.T) {
	c := newClient(t)
	token := c.authUser("t7@example.com")
	created := c.createTodo(token, fiber.Map{"title": "delete me"})
	var made todoView
	created.decodeData(t, &made)

	res := c.do(fiber.MethodDelete, fmt.Sprintf("/api/todos/%d", made.ID), token, nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("delete status = %d (%s)", res.status, res.raw)
	}

	// gone from list
	list := c.do(fiber.MethodGet, "/api/todos", token, nil)
	var page struct {
		Items []todoView `json:"items"`
	}
	list.decodeData(t, &page)
	if len(page.Items) != 0 {
		t.Errorf("soft-deleted todo still listed: %+v", page.Items)
	}

	// 404 on re-get
	res = c.do(fiber.MethodGet, fmt.Sprintf("/api/todos/%d", made.ID), token, nil)
	if res.status != fiber.StatusNotFound {
		t.Errorf("re-get status = %d, want 404", res.status)
	}

	// row still present with deleted_at set
	var deletedAt *string
	row := testDB.Raw("SELECT deleted_at FROM todos WHERE id = ?", made.ID).Row()
	if err := row.Scan(&deletedAt); err != nil {
		t.Fatalf("scan deleted_at: %v", err)
	}
	if deletedAt == nil {
		t.Error("expected deleted_at to be set (soft delete), row was hard-deleted")
	}

	// deleting again -> 404
	res = c.do(fiber.MethodDelete, fmt.Sprintf("/api/todos/%d", made.ID), token, nil)
	if res.status != fiber.StatusNotFound {
		t.Errorf("second delete status = %d, want 404", res.status)
	}
}

func TestTodo_OwnershipIsolation(t *testing.T) {
	c := newClient(t)
	tokenA := c.authUser("owner@example.com")
	tokenB := c.authUser("intruder@example.com")

	created := c.createTodo(tokenA, fiber.Map{"title": "A's secret"})
	var made todoView
	created.decodeData(t, &made)
	path := fmt.Sprintf("/api/todos/%d", made.ID)

	checks := []struct {
		method string
		url    string
		body   fiber.Map
	}{
		{fiber.MethodGet, path, nil},
		{fiber.MethodPut, path, fiber.Map{"title": "hacked"}},
		{fiber.MethodPatch, path + "/status", fiber.Map{"status": "completed"}},
		{fiber.MethodDelete, path, nil},
	}
	for _, ch := range checks {
		t.Run(ch.method, func(t *testing.T) {
			res := c.do(ch.method, ch.url, tokenB, ch.body)
			if res.status != fiber.StatusNotFound {
				t.Fatalf("%s by non-owner = %d, want 404 (%s)", ch.method, res.status, res.raw)
			}
		})
	}

	// A's todo is untouched
	res := c.do(fiber.MethodGet, path, tokenA, nil)
	var got todoView
	res.decodeData(t, &got)
	if got.Title != "A's secret" || got.Status != "pending" {
		t.Errorf("owner's todo was modified: %+v", got)
	}

	// B's list stays empty
	list := c.do(fiber.MethodGet, "/api/todos", tokenB, nil)
	var page struct {
		Items []todoView `json:"items"`
	}
	list.decodeData(t, &page)
	if len(page.Items) != 0 {
		t.Errorf("B should see no todos, got %+v", page.Items)
	}
}
