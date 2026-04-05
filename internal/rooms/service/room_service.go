package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
)

var ErrInvalidCapacity = errors.New("capacity must be greater than zero")

type RoomService interface {
	CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*model.Room, error)
	ListRooms(ctx context.Context) ([]*model.Room, error)
}

type roomService struct {
	repo repository.RoomRepository
}

func NewRoomService(r repository.RoomRepository) RoomService {
	return &roomService{repo: r}
}

func (s *roomService) CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*model.Room, error) {
	if req.Capacity != nil && *req.Capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	room := &model.Room{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
	}
	if err := s.repo.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) ListRooms(ctx context.Context) ([]*model.Room, error) {
	return s.repo.ListAll(ctx)
}
