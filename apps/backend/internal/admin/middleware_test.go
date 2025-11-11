package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	userModels "github.com/detect-price-by-photo/backend/internal/user/user/models"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

type stubUserRepo struct {
	user *userModels.User
	err  error
}

func (s *stubUserRepo) CreateUser(context.Context, *userModels.User) error { return nil }
func (s *stubUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*userModels.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}
func (s *stubUserRepo) ListUsers(context.Context, *userModels.FilterUser) ([]*userModels.User, error) {
	return nil, nil
}
func (s *stubUserRepo) UpdateUser(context.Context, *userModels.User) error   { return nil }
func (s *stubUserRepo) DeleteUser(context.Context, uuid.UUID) error          { return nil }
func (s *stubUserRepo) UsernameExists(context.Context, string) (bool, error) { return false, nil }
func (s *stubUserRepo) EmailExists(context.Context, string) (bool, error)    { return false, nil }
func (s *stubUserRepo) GetUserByEmail(context.Context, string) (*userModels.User, error) {
	return nil, nil
}
func (s *stubUserRepo) GetUserByUsername(context.Context, string) (*userModels.User, error) {
	return nil, nil
}

func TestRequireAdminAllowsAdmin(t *testing.T) {
	adminID := uuid.Must(uuid.NewV7())
	repo := &stubUserRepo{
		user: &userModels.User{ID: adminID, Role: "admin"},
	}

	mw := RequireAdmin(repo)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set("user_id", adminID.String())

	called := false
	handler := mw(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestRequireAdminBlocksNonAdmin(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	repo := &stubUserRepo{
		user: &userModels.User{ID: userID, Role: "user"},
	}

	mw := RequireAdmin(repo)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set("user_id", userID.String())

	err := mw(func(c echo.Context) error { return nil })(ctx)
	if err == nil {
		t.Fatalf("expected error for non-admin user")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", httpErr.Code)
	}
}
