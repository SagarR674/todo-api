package services

import (
	"errors"
	"testing"

	"github.com/SagarR674/todo-api/dto"
)

func TestCategoryService_Create(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepo())

	c, err := svc.Create(userA, dto.CreateCategoryRequest{Name: "  Work  "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.Name != "Work" {
		t.Errorf("name should be trimmed, got %q", c.Name)
	}
	if c.UserID != userA {
		t.Errorf("UserID = %d", c.UserID)
	}
}

func TestCategoryService_Create_Duplicate(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepo())
	_, _ = svc.Create(userA, dto.CreateCategoryRequest{Name: "Work"})

	_, err := svc.Create(userA, dto.CreateCategoryRequest{Name: "work"})
	if !errors.Is(err, ErrCategoryExists) {
		t.Fatalf("expected ErrCategoryExists, got %v", err)
	}
}

func TestCategoryService_Create_SameNameDifferentUser(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepo())
	_, _ = svc.Create(userA, dto.CreateCategoryRequest{Name: "Work"})

	if _, err := svc.Create(userB, dto.CreateCategoryRequest{Name: "Work"}); err != nil {
		t.Fatalf("another user should be able to reuse the name: %v", err)
	}
}

func TestCategoryService_Create_RepoErrorPropagates(t *testing.T) {
	repo := newFakeCategoryRepo()
	repo.createErr = errors.New("insert failed")
	svc := NewCategoryService(repo)

	if _, err := svc.Create(userA, dto.CreateCategoryRequest{Name: "Work"}); err == nil {
		t.Fatal("expected the repository Create error to propagate")
	}
}

func TestCategoryService_List(t *testing.T) {
	svc := NewCategoryService(newFakeCategoryRepo())
	_, _ = svc.Create(userA, dto.CreateCategoryRequest{Name: "Work"})
	_, _ = svc.Create(userA, dto.CreateCategoryRequest{Name: "Personal"})
	_, _ = svc.Create(userB, dto.CreateCategoryRequest{Name: "Other"})

	list, err := svc.List(userA)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("user A should see 2 categories, got %d", len(list))
	}
}
