package services

import (
	"errors"
	"testing"

	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/models"
)

const (
	userA uint = 1
	userB uint = 2
)

func TestTodoService_Create_Defaults(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)

	todo, err := svc.Create(userA, dto.CreateTodoRequest{Title: "  Ship it  "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if todo.Title != "Ship it" {
		t.Errorf("title should be trimmed, got %q", todo.Title)
	}
	if todo.Status != models.StatusPending {
		t.Errorf("default status = %q, want pending", todo.Status)
	}
	if todo.Priority != models.PriorityMedium {
		t.Errorf("default priority = %q, want medium", todo.Priority)
	}
	if todo.DueDate != nil {
		t.Errorf("due date should be nil by default, got %v", todo.DueDate)
	}
}

func TestTodoService_Create_WithDueDate(t *testing.T) {
	svc := NewTodoService(newFakeTodoRepo())

	todo, err := svc.Create(userA, dto.CreateTodoRequest{Title: "x", DueDate: "2026-09-15"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if todo.DueDate == nil || todo.DueDate.String() != "2026-09-15" {
		t.Fatalf("due date = %v", todo.DueDate)
	}

	if _, err := svc.Create(userA, dto.CreateTodoRequest{Title: "x", DueDate: "15/09/2026"}); !errors.Is(err, models.ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

func TestTodoService_Create_Categories(t *testing.T) {
	repo := newFakeTodoRepo()
	repo.seedCategory(1, userA, "Work")
	repo.seedCategory(2, userA, "Personal")
	repo.seedCategory(9, userB, "Secret")
	svc := NewTodoService(repo)

	todo, err := svc.Create(userA, dto.CreateTodoRequest{Title: "x", CategoryIDs: []uint{1, 2, 1}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(todo.Categories) != 2 {
		t.Errorf("duplicate category ids should be de-duplicated, got %d", len(todo.Categories))
	}

	// referencing another user's category -> not found
	if _, err := svc.Create(userA, dto.CreateTodoRequest{Title: "x", CategoryIDs: []uint{9}}); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
	// referencing a missing category -> not found
	if _, err := svc.Create(userA, dto.CreateTodoRequest{Title: "x", CategoryIDs: []uint{123}}); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestTodoService_Get_OwnershipEnforced(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)
	created, _ := svc.Create(userA, dto.CreateTodoRequest{Title: "mine"})

	if _, err := svc.Get(userA, created.ID); err != nil {
		t.Fatalf("owner should read their todo: %v", err)
	}
	if _, err := svc.Get(userB, created.ID); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("other user should get ErrTodoNotFound, got %v", err)
	}
	if _, err := svc.Get(userA, 9999); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("missing id should get ErrTodoNotFound, got %v", err)
	}
}

func TestTodoService_Update_Partial(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)
	created, _ := svc.Create(userA, dto.CreateTodoRequest{Title: "orig", Priority: "low"})

	newTitle := "updated"
	got, err := svc.Update(userA, created.ID, dto.UpdateTodoRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Title != "updated" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Priority != models.PriorityLow {
		t.Errorf("priority should be unchanged, got %q", got.Priority)
	}
}

func TestTodoService_Update_ClearDueDate(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)
	created, _ := svc.Create(userA, dto.CreateTodoRequest{Title: "x", DueDate: "2026-01-01"})

	empty := ""
	got, err := svc.Update(userA, created.ID, dto.UpdateTodoRequest{DueDate: &empty})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.DueDate != nil {
		t.Errorf("empty due_date should clear the field, got %v", got.DueDate)
	}
}

func TestTodoService_Update_NotOwner(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)
	created, _ := svc.Create(userA, dto.CreateTodoRequest{Title: "x"})

	title := "hax"
	if _, err := svc.Update(userB, created.ID, dto.UpdateTodoRequest{Title: &title}); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestTodoService_UpdateStatus(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)
	created, _ := svc.Create(userA, dto.CreateTodoRequest{Title: "x"})

	got, err := svc.UpdateStatus(userA, created.ID, models.StatusCompleted)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if got.Status != models.StatusCompleted {
		t.Errorf("status = %q", got.Status)
	}
	if _, err := svc.UpdateStatus(userB, created.ID, models.StatusCompleted); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("expected ErrTodoNotFound for non-owner, got %v", err)
	}
}

func TestTodoService_Delete(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)
	created, _ := svc.Create(userA, dto.CreateTodoRequest{Title: "x"})

	if err := svc.Delete(userB, created.ID); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("non-owner delete should fail, got %v", err)
	}
	if err := svc.Delete(userA, created.ID); err != nil {
		t.Fatalf("owner delete: %v", err)
	}
	if _, err := svc.Get(userA, created.ID); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("todo should be gone after delete, got %v", err)
	}
}

func TestTodoService_Create_RepoErrorPropagates(t *testing.T) {
	repo := newFakeTodoRepo()
	repo.createErr = errors.New("insert failed")
	svc := NewTodoService(repo)

	if _, err := svc.Create(userA, dto.CreateTodoRequest{Title: "x"}); err == nil {
		t.Fatal("expected the repository Create error to propagate")
	}
}

func TestTodoService_List(t *testing.T) {
	repo := newFakeTodoRepo()
	svc := NewTodoService(repo)
	_, _ = svc.Create(userA, dto.CreateTodoRequest{Title: "a"})
	_, _ = svc.Create(userA, dto.CreateTodoRequest{Title: "b"})
	_, _ = svc.Create(userB, dto.CreateTodoRequest{Title: "c"})

	page, err := svc.List(userA, &dto.ListTodosQuery{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Pagination.Total != 2 {
		t.Errorf("user A should see 2 todos, got %d", page.Pagination.Total)
	}
	if len(page.Items) != 2 {
		t.Errorf("items = %d", len(page.Items))
	}
}
