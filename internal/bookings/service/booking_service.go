package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/conference"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/dto"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/repository"
)

var (
	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotInPast        = errors.New("slot is in the past")
	ErrSlotAlreadyBooked = errors.New("slot is already booked")
	ErrConferenceFailed  = errors.New("conference link could not be created")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrCancelNotOwner    = errors.New("cannot cancel another user's booking")
)

type BookingService interface {
	ListMyFutureBookings(ctx context.Context, userID uuid.UUID) ([]*model.Booking, error)
	CreateBooking(ctx context.Context, userID uuid.UUID, req dto.CreateBookingRequest) (*model.Booking, error)
	CancelBooking(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error)
	ListAllBookingsPaged(ctx context.Context, page, pageSize int) (bookings []*model.Booking, total int64, err error)
}

type bookingService struct {
	repo       repository.BookingRepository
	conference conference.Client
}

func NewBookingService(r repository.BookingRepository, c conference.Client) BookingService {
	return &bookingService{repo: r, conference: c}
}

func (s *bookingService) ListMyFutureBookings(ctx context.Context, userID uuid.UUID) ([]*model.Booking, error) {
	return s.repo.ListByUserWithFutureSlots(ctx, userID, time.Now().UTC())
}

func (s *bookingService) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.CreateBookingRequest) (*model.Booking, error) {
	var confLink *string
	if req.CreateConferenceLink {
		url, err := s.conference.CreateLink(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrConferenceFailed, err)
		}
		confLink = &url
	}

	now := time.Now().UTC()
	b := &model.Booking{
		ID:             uuid.New(),
		SlotID:         req.SlotID,
		UserID:         userID,
		Status:         "active",
		ConferenceLink: confLink,
	}

	if err := s.repo.CreateActiveBookingWithSlotLock(ctx, b, req.SlotID, now); err != nil {
		switch {
		case errors.Is(err, repository.ErrBookingSlotNotFound):
			return nil, ErrSlotNotFound
		case errors.Is(err, repository.ErrBookingSlotInPast):
			return nil, ErrSlotInPast
		case errors.Is(err, repository.ErrActiveBookingExists):
			return nil, ErrSlotAlreadyBooked
		default:
			return nil, err
		}
	}
	return b, nil
}

func (s *bookingService) CancelBooking(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error) {
	b, err := s.repo.CancelBookingForUser(ctx, userID, bookingID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrCancelBookingNotFound):
			return nil, ErrBookingNotFound
		case errors.Is(err, repository.ErrCancelBookingForbidden):
			return nil, ErrCancelNotOwner
		default:
			return nil, err
		}
	}
	return b, nil
}

func (s *bookingService) ListAllBookingsPaged(ctx context.Context, page, pageSize int) ([]*model.Booking, int64, error) {
	return s.repo.ListAllPaged(ctx, page, pageSize)
}
