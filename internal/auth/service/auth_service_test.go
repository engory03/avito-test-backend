package service

import (
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
	byEmail func(string) (*usersModel.User, error)
	create  func(*usersModel.User) error
}

func (f *fakeAuthUserRepo) Create(u *usersModel.User) error {
	if f.create != nil {
		return f.create(u)
	}
	return nil
}
func (f *fakeAuthUserRepo) GetByID(uuid.UUID) (*usersModel.User, error) { return nil, errors.New("nf") }
func (f *fakeAuthUserRepo) GetByEmail(email string) (*usersModel.User, error) {
	if f.byEmail != nil {
		return f.byEmail(email)
	}
	return nil, errors.New("not found")
}
func (f *fakeAuthUserRepo) GetAll() ([]*usersModel.User, error) { return nil, nil }

type fakeRefreshRepo struct {
	save   func(*authModel.RefreshToken) error
	delete func(string) error
}

func (f *fakeRefreshRepo) Save(t *authModel.RefreshToken) error {
	if f.save != nil {
		return f.save(t)
	}
	return nil
}
func (f *fakeRefreshRepo) Delete(token string) error {
	if f.delete != nil {
		return f.delete(token)
	}
	return nil
}

func testJWT(t *testing.T) *utils.JWTManager {
	t.Helper()
	return utils.NewJWTManager("access-secret-test-key-32bytes!!", "refresh-secret-test-key-32bytes!")
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.DummyLogin(dto.DummyLoginRequest{Role: "guest"})
	if !errors.Is(err, ErrInvalidDummyRole) {
		t.Fatalf("got %v", err)
	}
}

func TestDummyLogin_AdminAndUser(t *testing.T) {
	j := testJWT(t)
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{}, j)
	adminTok, err := svc.DummyLogin(dto.DummyLoginRequest{Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	cl, err := j.VerifyAccessToken(adminTok)
	if err != nil || cl.Role != "admin" || cl.UserID != dummy.DummyAdminID.String() {
		t.Fatalf("claims %+v err %v", cl, err)
	}
	userTok, err := svc.DummyLogin(dto.DummyLoginRequest{Role: "user"})
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
		byEmail: func(string) (*usersModel.User, error) {
			return &usersModel.User{Email: "a@b.c"}, nil
		},
	}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.Register(dto.RegisterRequest{Email: "a@b.c", Password: "secret12", Role: "user"})
	if err == nil || err.Error() == "" {
		t.Fatalf("got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{
		byEmail: func(string) (*usersModel.User, error) {
			return nil, errors.New("not found")
		},
		create: func(*usersModel.User) error { return nil },
	}, &fakeRefreshRepo{}, testJWT(t))
	toks, err := svc.Register(dto.RegisterRequest{Email: "new@x.y", Password: "secret12", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	if toks.AccessToken == "" || toks.RefreshToken == "" {
		t.Fatal("tokens")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{
		byEmail: func(string) (*usersModel.User, error) {
			return nil, errors.New("not found")
		},
	}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.Login(dto.LoginRequest{Email: "x@y.z", Password: "wrong"})
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
		byEmail: func(email string) (*usersModel.User, error) {
			if email != u.Email {
				return nil, errors.New("nf")
			}
			return u, nil
		},
	}, &fakeRefreshRepo{}, testJWT(t))
	toks, err := svc.Login(dto.LoginRequest{Email: u.Email, Password: "correct-password"})
	if err != nil {
		t.Fatal(err)
	}
	if toks.AccessToken == "" {
		t.Fatal("no access")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{}, testJWT(t))
	_, err := svc.Refresh("not-a-jwt")
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
	// JWT `iat` is second-resolution; a new refresh token in the same second can equal the old one.
	time.Sleep(1100 * time.Millisecond)
	svc := NewAuthService(&fakeAuthUserRepo{}, &fakeRefreshRepo{
		save: func(*authModel.RefreshToken) error { return nil },
	}, j)
	toks, err := svc.Refresh(rt)
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
		delete: func(s string) error {
			called = true
			if s != "tok" {
				t.Fatal(s)
			}
			return nil
		},
	}, testJWT(t))
	if err := svc.Logout("tok"); err != nil || !called {
		t.Fatalf("called %v err %v", called, err)
	}
}
