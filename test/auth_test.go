package test

import (
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegister_Success(t *testing.T) {
	c := newClient(t)

	res := c.register("Alice", "alice@example.com", "password123")
	if res.status != fiber.StatusCreated {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	if !res.body.Success {
		t.Error("success should be true")
	}

	var user struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	res.decodeData(t, &user)
	if user.ID == 0 || user.Name != "Alice" || user.Email != "alice@example.com" {
		t.Errorf("unexpected user %+v", user)
	}

	// password / hash must never be returned
	if strings.Contains(string(res.raw), "password") {
		t.Errorf("response leaks password field: %s", res.raw)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	c := newClient(t)
	c.register("Alice", "dupe@example.com", "password123")

	res := c.register("Alice II", "DUPE@example.com", "password123")
	if res.status != fiber.StatusConflict {
		t.Fatalf("status = %d, want 409 (%s)", res.status, res.raw)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	c := newClient(t)

	cases := []struct {
		name  string
		body  fiber.Map
		field string
	}{
		{"missing name", fiber.Map{"email": "a@b.com", "password": "password123"}, "name"},
		{"bad email", fiber.Map{"name": "A", "email": "not-email", "password": "password123"}, "email"},
		{"short password", fiber.Map{"name": "A", "email": "a@b.com", "password": "short"}, "password"},
		{"empty body", fiber.Map{}, "name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := c.do(fiber.MethodPost, "/api/auth/register", "", tc.body)
			if res.status != fiber.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (%s)", res.status, res.raw)
			}
			if _, ok := res.body.Errors[tc.field]; !ok {
				t.Errorf("expected error for %q, got %v", tc.field, res.body.Errors)
			}
		})
	}
}

func TestRegister_MalformedJSON(t *testing.T) {
	c := newClient(t)
	res := c.doRaw(fiber.MethodPost, "/api/auth/register", "", `{"name": "A", "email":`)
	if res.status != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (%s)", res.status, res.raw)
	}
	if res.body.Success {
		t.Error("success should be false")
	}
}

func TestLogin_Success(t *testing.T) {
	c := newClient(t)
	c.register("Alice", "login@example.com", "password123")

	res := c.login("login@example.com", "password123")
	if res.status != fiber.StatusOK {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var out struct {
		Token string `json:"token"`
		User  struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	res.decodeData(t, &out)
	if out.Token == "" {
		t.Error("expected a token")
	}
	if out.User.Email != "login@example.com" {
		t.Errorf("user email = %q", out.User.Email)
	}
}

func TestLogin_Failures(t *testing.T) {
	c := newClient(t)
	c.register("Alice", "l2@example.com", "password123")

	cases := []struct {
		name string
		body fiber.Map
	}{
		{"wrong password", fiber.Map{"email": "l2@example.com", "password": "wrongpass"}},
		{"unknown user", fiber.Map{"email": "ghost@example.com", "password": "password123"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := c.do(fiber.MethodPost, "/api/auth/login", "", tc.body)
			if res.status != fiber.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 (%s)", res.status, res.raw)
			}
		})
	}
}

func TestMe_ReturnsCurrentUser(t *testing.T) {
	c := newClient(t)
	token := c.authUser("me@example.com")

	res := c.do(fiber.MethodGet, "/api/auth/me", token, nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var user struct {
		Email string `json:"email"`
	}
	res.decodeData(t, &user)
	if user.Email != "me@example.com" {
		t.Errorf("email = %q", user.Email)
	}
	if strings.Contains(string(res.raw), "password") {
		t.Errorf("/me leaks password: %s", res.raw)
	}
}

func TestMe_RequiresAuth(t *testing.T) {
	c := newClient(t)
	res := c.do(fiber.MethodGet, "/api/auth/me", "", nil)
	if res.status != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.status)
	}
}
