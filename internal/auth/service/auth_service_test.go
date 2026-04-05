package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dto"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
	authModel "github.com/internships-backend/test-backend-engory03/internal/auth/model"
	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	usersModel "github.com/internships-backend/test-backend-engory03/internal/users/model"
)

type fakeAuthUserRepo struct {
	byEmail func(context.Context, string) (*usersModel.User, error)
	create  func(context.Context, *usersModel.User) error
}

func (f *fakeAuthUserRepo) Create(ctx context.Context, u *usersModel.User) error {
	if f.create != nil {
		return f.create(ctx, u)
	}
	return nil
}
func (f *fakeAuthUserRepo) GetByID(context.Context, uuid.UUID) (*usersModel.User, error) {
	return nil, errors.New("nf")
}
func (f *fakeAuthUserRepo) GetByEmail(ctx context.Context, email string) (*usersModel.User, error) {
	if f.byEmail != nil {
		return f.byEmail(ctx, email)
	}
	return nil, errors.New("not found")
}
func (f *fakeAuthUserRepo) GetAll(context.Context) ([]*usersModel.User, error) { return nil, nil }

type fakeRefreshRepo struct {
	save   func(context.Context, *authModel.RefreshToken) error
	delete func(context.Context, string) error
}

func (f *fakeRefreshRepo) Save(ctx context.Context, t *authModel.RefreshToken) error {
	if f.save != nil {
		return f.save(ctx, t)
	}
	return nil
}
func (f *fakeRefreshRepo) Delete(ctx context.Context, token string) error {
	if f.delete != nil {
		return f.delete(ctx, token)
	}
	return nil
}

func testJWT(t *testing.T) *utils.JWTManager {
	t.Helper()
	return utils.NewJWTManager("access-secret-test-key-32bytes!!", "refresh-secret-test-key-32bytes!")
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.DummyLogin(context.Background(), dto.DummyLoginRequest{Role: "guest"})
	if !errors.Is(err, ErrInvalidDummyRole) {
		t.Fatalf("got %v", err)
	}
}

func TestDummyLogin_AdminAndUser(t *testing.T) {
	j := testJWT(t)
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{}, j)
	adminTok, err := svc.DummyLogin(context.Background(), dto.DummyLoginRequest{Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	cl, err := j.VerifyAccessToken(adminTok)
	if err != nil || cl.Role != "admin" || cl.UserID != dummy.DummyAdminID.String() {
		t.Fatalf("claims %+v err %v", cl, err)
	}
	userTok, err := svc.DummyLogin(context.Background(), dto.DummyLoginRequest{Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	cl2, err := j.VerifyAccessToken(userTok)
	if err != nil || cl2.UserID != dummy.DummyUserID.String() {
		t.Fatalf("user claims %+v", cl2)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{
		byEmail: func(context.Context, string) (*usersModel.User, error) {
			return &usersModel.User{Email: "a@b.c"}, nil
		},
	}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.Register(context.Background(), dto.RegisterRequest{Email: "a@b.c", Password: "secret12", Role: "user"})
	if err == nil || err.Error() == "" {
		t.Fatalf("got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{
		byEmail: func(context.Context, string) (*usersModel.User, error) {
			return nil, errors.New("not found")
		},
		create: func(context.Context, *usersModel.User) error { return nil },
	}, &fakeRefreshRepo{}, testJWT(t))
	toks, err := svc.Register(context.Background(), dto.RegisterRequest{Email: "new@x.y", Password: "secret12", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	if toks.AccessToken == "" || toks.RefreshToken == "" {
		t.Fatal("tokens")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{
		byEmail: func(context.Context, string) (*usersModel.User, error) {
			return nil, errors.New("not found")
		},
	}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.Login(context.Background(), dto.LoginRequest{Email: "x@y.z", Password: "wrong"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLogin_Success(t *testing.T) {
	hash, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	u := &usersModel.User{
		ID: uuid.New(), Email: "ok@x.y", PasswordHash: hash, Role: "user",
	}
	svc := NewAuthService(&fakeAuthUserRepo{
		byEmail: func(_ context.Context, email string) (*usersModel.User, error) {
			if email != u.Email {
				return nil, errors.New("nf")
			}
			return u, nil
		},
	}, &fakeRefreshRepo{}, testJWT(t))
	toks, err := svc.Login(context.Background(), dto.LoginRequest{Email: u.Email, Password: "correct-password"})
	if err != nil {
		t.Fatal(err)
	}
	if toks.AccessToken == "" {
		t.Fatal("no access")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.Refresh(context.Background(), "not-a-jwt")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRefresh_Success(t *testing.T) {
	j := testJWT(t)
	uid := uuid.New()
	rt, err := j.GenerateRefreshToken(uid, "user")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{
		save: func(context.Context, *authModel.RefreshToken) error { return nil },
	}, j)
	toks, err := svc.Refresh(context.Background(), rt)
	if err != nil {
		t.Fatal(err)
	}
	if toks.AccessToken == "" || toks.RefreshToken == "" {
		t.Fatalf("empty tokens")
	}
	if toks.RefreshToken == rt {
		t.Fatalf("refresh token should rotate")
	}
}

func TestLogout(t *testing.T) {
	called := false
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{
		delete: func(_ context.Context, s string) error {
			called = true
			if s != "tok" {
				t.Fatal(s)
			}
			return nil
		},
	}, testJWT(t))
	if err := svc.Logout(context.Background(), "tok"); err != nil || !called {
		t.Fatalf("called %v err %v", called, err)
	}
}
