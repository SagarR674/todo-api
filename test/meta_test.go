package test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestHealth(t *testing.T) {
	c := newClient(t)
	res := c.do(fiber.MethodGet, "/health", "", nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	var data struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}
	res.decodeData(t, &data)
	if data.Status != "ok" || data.Database != "connected" {
		t.Errorf("health data = %+v", data)
	}
}

func TestIndexRoute(t *testing.T) {
	c := newClient(t)
	res := c.do(fiber.MethodGet, "/", "", nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("status = %d (%s)", res.status, res.raw)
	}
	if !res.body.Success {
		t.Error("index route should return success=true")
	}
}

func TestUnknownRoute_Returns404Envelope(t *testing.T) {
	c := newClient(t)
	res := c.do(fiber.MethodGet, "/no/such/path", "", nil)
	if res.status != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.status)
	}
	if res.body.Success || res.body.Message == "" {
		t.Errorf("expected error envelope, got %+v", res.body)
	}
}
