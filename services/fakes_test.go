package services

import (
	"strings"

	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/models"
	"github.com/SagarR674/todo-api/repository"
)

// ---- fake user repository -------------------------------------------------

type fakeUserRepo struct {
	byEmail   map[string]*models.User
	nextID    uint
	createErr error
	existsErr error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]*models.User{}, nextID: 1}
}

func (f *fakeUserRepo) Create(u *models.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	u.ID = f.nextID
	f.nextID++
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) FindByEmail(email string) (*models.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func (f *fakeUserRepo) FindByID(id uint) (*models.User, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *fakeUserRepo) ExistsByEmail(email string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	_, ok := f.byEmail[email]
	return ok, nil
}

// ---- fake token issuer --------------------------------------------------

type fakeTokenIssuer struct {
	err error
}

func (f fakeTokenIssuer) Generate(userID uint) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "token-for-user", nil
}

// ---- fake todo repository ---------------------------------------------

type fakeTodoRepo struct {
	todos      map[uint]*models.Todo
	categories map[uint]models.Category // categoryID -> category (owned check via UserID)
	nextID     uint
	createErr  error
}

func newFakeTodoRepo() *fakeTodoRepo {
	return &fakeTodoRepo{
		todos:      map[uint]*models.Todo{},
		categories: map[uint]models.Category{},
		nextID:     1,
	}
}

func (f *fakeTodoRepo) seedCategory(id, userID uint, name string) {
	f.categories[id] = models.Category{ID: id, UserID: userID, Name: name}
}

func (f *fakeTodoRepo) Create(todo *models.Todo) error {
	if f.createErr != nil {
		return f.createErr
	}
	todo.ID = f.nextID
	f.nextID++
	cp := *todo
	f.todos[todo.ID] = &cp
	return nil
}

func (f *fakeTodoRepo) FindByIDForUser(id, userID uint) (*models.Todo, error) {
	if t, ok := f.todos[id]; ok && t.UserID == userID {
		cp := *t
		return &cp, nil
	}
	return nil, repository.ErrNotFound
}

func (f *fakeTodoRepo) List(userID uint, _ *dto.ListTodosQuery) ([]models.Todo, int64, error) {
	var out []models.Todo
	for _, t := range f.todos {
		if t.UserID == userID {
			out = append(out, *t)
		}
	}
	return out, int64(len(out)), nil
}

func (f *fakeTodoRepo) Save(todo *models.Todo) error {
	if _, ok := f.todos[todo.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *todo
	f.todos[todo.ID] = &cp
	return nil
}

func (f *fakeTodoRepo) UpdateStatus(id, userID uint, status models.Status) error {
	if t, ok := f.todos[id]; ok && t.UserID == userID {
		t.Status = status
		return nil
	}
	return repository.ErrNotFound
}

func (f *fakeTodoRepo) ReplaceCategories(todo *models.Todo, categories []models.Category) error {
	if t, ok := f.todos[todo.ID]; ok {
		t.Categories = categories
		return nil
	}
	return repository.ErrNotFound
}

func (f *fakeTodoRepo) SoftDelete(id, userID uint) error {
	if t, ok := f.todos[id]; ok && t.UserID == userID {
		delete(f.todos, id)
		return nil
	}
	return repository.ErrNotFound
}

func (f *fakeTodoRepo) CategoriesForUser(userID uint, ids []uint) ([]models.Category, error) {
	var out []models.Category
	for _, id := range ids {
		if c, ok := f.categories[id]; ok && c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

// ---- fake category repository ----------------------------------------

type fakeCategoryRepo struct {
	items     []models.Category
	nextID    uint
	createErr error
}

func newFakeCategoryRepo() *fakeCategoryRepo {
	return &fakeCategoryRepo{nextID: 1}
}

func (f *fakeCategoryRepo) Create(c *models.Category) error {
	if f.createErr != nil {
		return f.createErr
	}
	for _, existing := range f.items {
		if existing.UserID == c.UserID && strings.EqualFold(existing.Name, c.Name) {
			return repository.ErrDuplicate
		}
	}
	c.ID = f.nextID
	f.nextID++
	f.items = append(f.items, *c)
	return nil
}

func (f *fakeCategoryRepo) ListForUser(userID uint) ([]models.Category, error) {
	var out []models.Category
	for _, c := range f.items {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

// compile-time checks that the fakes satisfy the service interfaces
var (
	_ UserRepo     = (*fakeUserRepo)(nil)
	_ TokenIssuer  = fakeTokenIssuer{}
	_ TodoRepo     = (*fakeTodoRepo)(nil)
	_ CategoryRepo = (*fakeCategoryRepo)(nil)
)
