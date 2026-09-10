package services

import (
	"errors"
	"testing"

	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/utils"
)

func TestAuthService_Register(t *testing.T) {
	users := newFakeUserRepo()
	svc := NewAuthService(users, fakeTokenIssuer{})

	user, err := svc.Register(dto.RegisterRequest{
		Name: "  Alice  ", Email: "Alice@Example.com ", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Name != "Alice" {
		t.Errorf("name should be trimmed, got %q", user.Name)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("email should be normalised, got %q", user.Email)
	}
	if user.Password == "password123" {
		t.Error("password must be hashed")
	}
	if !utils.CheckPassword(user.Password, "password123") {
		t.Error("stored hash should verify against the original password")
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	users := newFakeUserRepo()
	svc := NewAuthService(users, fakeTokenIssuer{})

	_, _ = svc.Register(dto.RegisterRequest{Name: "A", Email: "a@x.com", Password: "password123"})
	_, err := svc.Register(dto.RegisterRequest{Name: "B", Email: "a@x.com", Password: "password123"})
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	users := newFakeUserRepo()
	svc := NewAuthService(users, fakeTokenIssuer{})
	_, _ = svc.Register(dto.RegisterRequest{Name: "A", Email: "a@x.com", Password: "password123"})

	token, user, err := svc.Login(dto.LoginRequest{Email: "A@X.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" || user == nil {
		t.Fatalf("expected token and user, got %q / %v", token, user)
	}
}

func TestAuthService_Login_Failures(t *testing.T) {
	users := newFakeUserRepo()
	svc := NewAuthService(users, fakeTokenIssuer{})
	_, _ = svc.Register(dto.RegisterRequest{Name: "A", Email: "a@x.com", Password: "password123"})

	tests := []struct {
		name string
		req  dto.LoginRequest
	}{
		{"wrong password", dto.LoginRequest{Email: "a@x.com", Password: "nope"}},
		{"unknown user", dto.LoginRequest{Email: "ghost@x.com", Password: "password123"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Login(tc.req)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("expected ErrInvalidCredentials, got %v", err)
			}
		})
	}
}

func TestAuthService_Register_RepoErrorsPropagate(t *testing.T) {
	users := newFakeUserRepo()
	users.existsErr = errors.New("db unreachable")
	svc := NewAuthService(users, fakeTokenIssuer{})

	if _, err := svc.Register(dto.RegisterRequest{Name: "A", Email: "a@x.com", Password: "password123"}); err == nil {
		t.Fatal("expected the ExistsByEmail error to propagate")
	}

	users2 := newFakeUserRepo()
	users2.createErr = errors.New("insert failed")
	svc2 := NewAuthService(users2, fakeTokenIssuer{})
	if _, err := svc2.Register(dto.RegisterRequest{Name: "A", Email: "a@x.com", Password: "password123"}); err == nil {
		t.Fatal("expected the Create error to propagate")
	}
}

func TestAuthService_Profile(t *testing.T) {
	users := newFakeUserRepo()
	svc := NewAuthService(users, fakeTokenIssuer{})
	created, _ := svc.Register(dto.RegisterRequest{Name: "A", Email: "a@x.com", Password: "password123"})

	got, err := svc.Profile(created.ID)
	if err != nil {
		t.Fatalf("Profile: %v", err)
	}
	if got.Email != "a@x.com" {
		t.Errorf("email = %q", got.Email)
	}

	if _, err := svc.Profile(9999); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unknown user should map to ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_TokenIssuerError(t *testing.T) {
	users := newFakeUserRepo()
	svc := NewAuthService(users, fakeTokenIssuer{err: errors.New("kms down")})
	_, _ = svc.Register(dto.RegisterRequest{Name: "A", Email: "a@x.com", Password: "password123"})

	if _, _, err := svc.Login(dto.LoginRequest{Email: "a@x.com", Password: "password123"}); err == nil {
		t.Fatal("expected the token issuer error to propagate")
	}
}
