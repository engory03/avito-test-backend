package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/users/dto"
	"github.com/internships-backend/test-backend-engory03/internal/users/model"
)

type fakeUsersRepo struct {
	byEmail func(context.Context, string) (*model.User, error)
	create  func(context.Context, *model.User) error
	byID    func(context.Context, uuid.UUID) (*model.User, error)
	all     func(context.Context) ([]*model.User, error)
}

func (f *fakeUsersRepo) Create(ctx context.Context, u *model.User) error {
	if f.create != nil {
		return f.create(ctx, u)
	}
	return nil
}
func (f *fakeUsersRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if f.byID != nil {
		return f.byID(ctx, id)
	}
	return nil, errors.New("nf")
}
func (f *fakeUsersRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if f.byEmail != nil {
		return f.byEmail(ctx, email)
	}
	return nil, errors.New("nf")
}
func (f *fakeUsersRepo) GetAll(ctx context.Context) ([]*model.User, error) {
	if f.all != nil {
		return f.all(ctx)
	}
	return nil, nil
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	svc := NewUserService(&fakeUsersRepo{
		byEmail: func(context.Context, string) (*model.User, error) {
			return &model.User{}, nil
		},
	})
	_, err := svc.CreateUser(context.Background(), dto.CreateUserRequest{Email: "a@b.c", Password: "longenough", Role: "user"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateUser_Success(t *testing.T) {
	var saved *model.User
	svc := NewUserService(&fakeUsersRepo{
		byEmail: func(context.Context, string) (*model.User, error) {
			return nil, errors.New("nf")
		},
		create: func(_ context.Context, u *model.User) error {
			saved = u
			return nil
		},
	})
	got, err := svc.CreateUser(context.Background(), dto.CreateUserRequest{
		Email: "u@x.y", Password: "password1", Role: "user",
	})
	if err != nil || saved == nil || got.Email != "u@x.y" {
		t.Fatalf("got %+v saved %+v err %v", got, saved, err)
	}
}

func TestGetUserById_NotFound(t *testing.T) {
	svc := NewUserService(&fakeUsersRepo{
		byID: func(context.Context, uuid.UUID) (*model.User, error) {
			return nil, errors.New("nf")
		},
	})
	_, err := svc.GetUserById(context.Background(), uuid.New())
	if err == nil || err.Error() != "user not found" {
		t.Fatalf("got %v", err)
	}
}

func TestGetUserById_Success(t *testing.T) {
	id := uuid.New()
	want := &model.User{ID: id, Email: "a@b.c", Role: "user"}
	svc := NewUserService(&fakeUsersRepo{
		byID: func(_ context.Context, got uuid.UUID) (*model.User, error) {
			if got != id {
				t.Fatalf("id %v", got)
			}
			return want, nil
		},
	})
	u, err := svc.GetUserById(context.Background(), id)
	if err != nil || u.Email != want.Email {
		t.Fatalf("got %+v err %v", u, err)
	}
}

func TestGetByEmail_Delegates(t *testing.T) {
	want := &model.User{ID: uuid.New(), Email: "x@y.z"}
	svc := NewUserService(&fakeUsersRepo{
		byEmail: func(_ context.Context, email string) (*model.User, error) {
			if email != "x@y.z" {
				t.Fatal(email)
			}
			return want, nil
		},
	})
	u, err := svc.GetByEmail(context.Background(), "x@y.z")
	if err != nil || u.ID != want.ID {
		t.Fatalf("got %+v err %v", u, err)
	}
}

func TestGetAllUsers(t *testing.T) {
	want := []*model.User{{ID: uuid.New(), Email: "e"}}
	svc := NewUserService(&fakeUsersRepo{
		all: func(context.Context) ([]*model.User, error) {
			return want, nil
		},
	})
	got, err := svc.GetAllUsers(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
}
