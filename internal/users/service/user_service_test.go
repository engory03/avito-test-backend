package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/users/dto"
	"github.com/internships-backend/test-backend-engory03/internal/users/model"
)

type fakeUsersRepo struct {
	byEmail func(string) (*model.User, error)
	create  func(*model.User) error
	byID    func(uuid.UUID) (*model.User, error)
	all     func() ([]*model.User, error)
}

func (f *fakeUsersRepo) Create(u *model.User) error {
	if f.create != nil {
		return f.create(u)
	}
	return nil
}
func (f *fakeUsersRepo) GetByID(id uuid.UUID) (*model.User, error) {
	if f.byID != nil {
		return f.byID(id)
	}
	return nil, errors.New("nf")
}
func (f *fakeUsersRepo) GetByEmail(email string) (*model.User, error) {
	if f.byEmail != nil {
		return f.byEmail(email)
	}
	return nil, errors.New("nf")
}
func (f *fakeUsersRepo) GetAll() ([]*model.User, error) {
	if f.all != nil {
		return f.all()
	}
	return nil, nil
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	svc := NewUserService(&fakeUsersRepo{
		byEmail: func(string) (*model.User, error) {
			return &model.User{}, nil
		},
	})
	_, err := svc.CreateUser(dto.CreateUserRequest{Email: "a@b.c", Password: "longenough", Role: "user"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateUser_Success(t *testing.T) {
	var saved *model.User
	svc := NewUserService(&fakeUsersRepo{
		byEmail: func(string) (*model.User, error) {
			return nil, errors.New("nf")
		},
		create: func(u *model.User) error {
			saved = u
			return nil
		},
	})
	got, err := svc.CreateUser(dto.CreateUserRequest{
		Email: "u@x.y", Password: "password1", Role: "user",
	})
	if err != nil || saved == nil || got.Email != "u@x.y" {
		t.Fatalf("got %+v saved %+v err %v", got, saved, err)
	}
}

func TestGetUserById_NotFound(t *testing.T) {
	svc := NewUserService(&fakeUsersRepo{
		byID: func(_ uuid.UUID) (*model.User, error) {
			return nil, errors.New("nf")
		},
	})
	_, err := svc.GetUserById(uuid.New())
	if err == nil || err.Error() != "user not found" {
		t.Fatalf("got %v", err)
	}
}

func TestGetAllUsers(t *testing.T) {
	want := []*model.User{{ID: uuid.New(), Email: "e"}}
	svc := NewUserService(&fakeUsersRepo{
		all: func() ([]*model.User, error) {
			return want, nil
		},
	})
	got, err := svc.GetAllUsers()
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
}
