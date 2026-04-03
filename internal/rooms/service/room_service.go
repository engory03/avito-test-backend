package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
)

var ErrInvalidCapacity = errors.New("capacity must be greater than zero")

type RoomService interface {
	CreateRoom(req dto.CreateRoomRequest) (*model.Room, error)
	ListRooms() ([]*model.Room, error)
}

type roomService struct {
	repo repository.RoomRepository
}

func NewRoomService(r repository.RoomRepository) RoomService {
	return &roomService{repo: r}
}

func (s *roomService) CreateRoom(req dto.CreateRoomRequest) (*model.Room, error) {
	if req.Capacity != nil && *req.Capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	room := &model.Room{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
	}
	if err := s.repo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) ListRooms() ([]*model.Room, error) {
	return s.repo.ListAll()
}
