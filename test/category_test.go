package test

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCategory_CreateAndList(t *testing.T) {
	c := newClient(t)
	token := c.authUser("cat@example.com")

	for _, name := range []string{"Work", "Personal", "Learning"} {
		res := c.do(fiber.MethodPost, "/api/categories", token, fiber.Map{"name": name})
		if res.status != fiber.StatusCreated {
			t.Fatalf("create %s: %d (%s)", name, res.status, res.raw)
		}
	}

	res := c.do(fiber.MethodGet, "/api/categories", token, nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("list: %d (%s)", res.status, res.raw)
	}
	var cats []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	res.decodeData(t, &cats)
	if len(cats) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(cats))
	}
}

func TestCategory_DuplicateRejected(t *testing.T) {
	c := newClient(t)
	token := c.authUser("cat2@example.com")

	c.do(fiber.MethodPost, "/api/categories", token, fiber.Map{"name": "Work"})
	res := c.do(fiber.MethodPost, "/api/categories", token, fiber.Map{"name": "work"})
	if res.status != fiber.StatusConflict {
		t.Fatalf("duplicate category status = %d, want 409 (%s)", res.status, res.raw)
	}
}

func TestCategory_Validation(t *testing.T) {
	c := newClient(t)
	token := c.authUser("cat3@example.com")

	res := c.do(fiber.MethodPost, "/api/categories", token, fiber.Map{"name": ""})
	if res.status != fiber.StatusBadRequest {
		t.Fatalf("empty name status = %d, want 400", res.status)
	}
}

func TestCategory_RequiresAuth(t *testing.T) {
	c := newClient(t)
	res := c.do(fiber.MethodGet, "/api/categories", "", nil)
	if res.status != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.status)
	}
}

func TestTodo_WithCategories(t *testing.T) {
	c := newClient(t)
	token := c.authUser("catodo@example.com")

	mk := func(name string) uint {
		res := c.do(fiber.MethodPost, "/api/categories", token, fiber.Map{"name": name})
		var cat struct {
			ID uint `json:"id"`
		}
		res.decodeData(t, &cat)
		return cat.ID
	}
	work := mk("Work")
	learning := mk("Learning")

	res := c.createTodo(token, fiber.Map{"title": "study", "category_ids": []uint{work, learning}})
	if res.status != fiber.StatusCreated {
		t.Fatalf("create: %d (%s)", res.status, res.raw)
	}
	var todo todoView
	res.decodeData(t, &todo)
	if len(todo.Categories) != 2 {
		t.Fatalf("expected 2 categories on todo, got %d", len(todo.Categories))
	}

	// filter todos by category name
	p := list(t, c, token, "category=Work&limit=100")
	if p.Pagination.Total != 1 {
		t.Errorf("category filter total = %d, want 1", p.Pagination.Total)
	}

	// referencing another user's category id -> 400
	other := c.authUser("catodo-other@example.com")
	otherCatRes := c.do(fiber.MethodPost, "/api/categories", other, fiber.Map{"name": "Secret"})
	var otherCat struct {
		ID uint `json:"id"`
	}
	otherCatRes.decodeData(t, &otherCat)

	res = c.createTodo(token, fiber.Map{"title": "x", "category_ids": []uint{otherCat.ID}})
	if res.status != fiber.StatusBadRequest {
		t.Errorf("cross-user category status = %d, want 400 (%s)", res.status, res.raw)
	}

	// removing categories via update
	upd := c.do(fiber.MethodPut, fmt.Sprintf("/api/todos/%d", todo.ID), token, fiber.Map{"category_ids": []uint{}})
	var updated todoView
	upd.decodeData(t, &updated)
	if len(updated.Categories) != 0 {
		t.Errorf("categories should be cleared, got %d", len(updated.Categories))
	}
}
