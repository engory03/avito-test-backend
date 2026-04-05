package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/dto"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
	"github.com/lib/pq"
)

var (
	ErrRoomNotFound   = errors.New("room not found")
	ErrScheduleExists = errors.New("schedule for this room already exists and cannot be changed")
)

type ScheduleService interface {
	CreateSchedule(ctx context.Context, pathRoomID uuid.UUID, req dto.ScheduleCreateRequest) (*model.Schedule, error)
}

type scheduleService struct {
	roomRepo     repository.RoomRepository
	scheduleRepo repository.ScheduleRepository
}

func NewScheduleService(roomRepo repository.RoomRepository, scheduleRepo repository.ScheduleRepository) ScheduleService {
	return &scheduleService{
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *scheduleService) CreateSchedule(ctx context.Context, pathRoomID uuid.UUID, req dto.ScheduleCreateRequest) (*model.Schedule, error) {
	if req.RoomID != pathRoomID {
		return nil, &ScheduleValidationError{Message: "roomId in path must match roomId in body"}
	}
	if err := validateSchedulePayload(&req); err != nil {
		return nil, err
	}
	ok, err := s.roomRepo.ExistsByID(ctx, pathRoomID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrRoomNotFound
	}
	has, err := s.scheduleRepo.ExistsByRoomID(ctx, pathRoomID)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, ErrScheduleExists
	}
	days := append([]int(nil), req.DaysOfWeek...)
	daysPQ := make(pq.Int64Array, len(days))
	for i, d := range days {
		daysPQ[i] = int64(d)
	}
	sched := &model.Schedule{
		ID:         uuid.New(),
		RoomID:     pathRoomID,
		DaysOfWeek: daysPQ,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	}
	if err := s.scheduleRepo.Create(ctx, sched); err != nil {
		return nil, err
	}
	return sched, nil
}
