package utils

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func doRequest(t *testing.T, handler fiber.Handler) (int, Response) {
	t.Helper()
	app := fiber.New()
	app.Get("/", handler)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var out Response
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal %s: %v", body, err)
	}
	return resp.StatusCode, out
}

func TestSuccess(t *testing.T) {
	code, body := doRequest(t, func(c *fiber.Ctx) error {
		return Success(c, fiber.StatusCreated, "done", fiber.Map{"id": 1})
	})
	if code != fiber.StatusCreated {
		t.Errorf("status = %d, want 201", code)
	}
	if !body.Success || body.Message != "done" {
		t.Errorf("body = %+v", body)
	}
	if body.Data == nil {
		t.Error("expected data payload")
	}
}

func TestError(t *testing.T) {
	code, body := doRequest(t, func(c *fiber.Ctx) error {
		return Error(c, fiber.StatusNotFound, "missing")
	})
	if code != fiber.StatusNotFound {
		t.Errorf("status = %d, want 404", code)
	}
	if body.Success || body.Message != "missing" || body.Data != nil {
		t.Errorf("body = %+v", body)
	}
}

func TestValidationError(t *testing.T) {
	code, body := doRequest(t, func(c *fiber.Ctx) error {
		return ValidationError(c, map[string]string{"title": "title is required"})
	})
	if code != fiber.StatusBadRequest {
		t.Errorf("status = %d, want 400", code)
	}
	if body.Success || body.Errors == nil {
		t.Errorf("body = %+v", body)
	}
}
