package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/conference"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/dto"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/repository"
	roomsModel "github.com/internships-backend/test-backend-engory03/internal/rooms/model"
)

type fakeBookingRepo struct {
	createWithSlotLock func(context.Context, *model.Booking, uuid.UUID, time.Time) error
	cancelForUser      func(context.Context, uuid.UUID, uuid.UUID) (*model.Booking, error)
	getByID            func(context.Context, uuid.UUID) (*model.Booking, error)
	listMy             func(context.Context, uuid.UUID, time.Time) ([]*model.Booking, error)
	listAllPaged       func(context.Context, int, int) ([]*model.Booking, int64, error)
	lastGetByIDResult  *model.Booking
}

func (f *fakeBookingRepo) ListByUserWithFutureSlots(ctx context.Context, userID uuid.UUID, nowUTC time.Time) ([]*model.Booking, error) {
	if f.listMy != nil {
		return f.listMy(ctx, userID, nowUTC)
	}
	return nil, nil
}

func (f *fakeBookingRepo) GetSlotByID(context.Context, uuid.UUID) (*roomsModel.Slot, error) {
	return nil, nil
}

func (f *fakeBookingRepo) CreateActiveBookingWithSlotLock(ctx context.Context, booking *model.Booking, slotID uuid.UUID, nowUTC time.Time) error {
	if f.createWithSlotLock != nil {
		return f.createWithSlotLock(ctx, booking, slotID, nowUTC)
	}
	return nil
}

func (f *fakeBookingRepo) CancelBookingForUser(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error) {
	if f.cancelForUser != nil {
		return f.cancelForUser(ctx, userID, bookingID)
	}
	return nil, errors.New("noop")
}

func (f *fakeBookingRepo) GetByID(ctx context.Context, bookingID uuid.UUID) (*model.Booking, error) {
	if f.getByID != nil {
		b, err := f.getByID(ctx, bookingID)
		f.lastGetByIDResult = b
		return b, err
	}
	return nil, nil
}

func (f *fakeBookingRepo) ListAllPaged(ctx context.Context, page, pageSize int) ([]*model.Booking, int64, error) {
	if f.listAllPaged != nil {
		return f.listAllPaged(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func TestListMyFutureBookings_DelegatesToRepo(t *testing.T) {
	uid := uuid.New()
	want := []*model.Booking{{ID: uuid.New(), UserID: uid}}
	repo := &fakeBookingRepo{
		listMy: func(_ context.Context, u uuid.UUID, n time.Time) ([]*model.Booking, error) {
			if u != uid {
				t.Fatalf("user id mismatch")
			}
			return want, nil
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	got, err := svc.ListMyFutureBookings(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestCreateBooking_SlotNotFound(t *testing.T) {
	slotID := uuid.New()
	repo := &fakeBookingRepo{
		createWithSlotLock: func(context.Context, *model.Booking, uuid.UUID, time.Time) error {
			return repository.ErrBookingSlotNotFound
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	_, err := svc.CreateBooking(context.Background(), uuid.New(), dto.CreateBookingRequest{SlotID: slotID})
	if !errors.Is(err, ErrSlotNotFound) {
		t.Fatalf("got %v want ErrSlotNotFound", err)
	}
}

func TestCreateBooking_SlotInPast(t *testing.T) {
	slotID := uuid.New()
	repo := &fakeBookingRepo{
		createWithSlotLock: func(context.Context, *model.Booking, uuid.UUID, time.Time) error {
			return repository.ErrBookingSlotInPast
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	_, err := svc.CreateBooking(context.Background(), uuid.New(), dto.CreateBookingRequest{SlotID: slotID})
	if !errors.Is(err, ErrSlotInPast) {
		t.Fatalf("got %v want ErrSlotInPast", err)
	}
}

func TestCreateBooking_Success(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()
	repo := &fakeBookingRepo{
		createWithSlotLock: func(_ context.Context, b *model.Booking, sid uuid.UUID, _ time.Time) error {
			if sid != slotID || b.UserID != userID || b.Status != "active" {
				t.Fatalf("booking fields")
			}
			return nil
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	b, err := svc.CreateBooking(context.Background(), userID, dto.CreateBookingRequest{SlotID: slotID})
	if err != nil {
		t.Fatal(err)
	}
	if b.SlotID != slotID || b.UserID != userID {
		t.Fatalf("returned booking %+v", b)
	}
}

func TestCreateBooking_SlotAlreadyBooked(t *testing.T) {
	slotID := uuid.New()
	repo := &fakeBookingRepo{
		createWithSlotLock: func(context.Context, *model.Booking, uuid.UUID, time.Time) error {
			return repository.ErrActiveBookingExists
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	_, err := svc.CreateBooking(context.Background(), uuid.New(), dto.CreateBookingRequest{SlotID: slotID})
	if !errors.Is(err, ErrSlotAlreadyBooked) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateBooking_ConferenceLink(t *testing.T) {
	slotID := uuid.New()
	repo := &fakeBookingRepo{
		createWithSlotLock: func(_ context.Context, b *model.Booking, _ uuid.UUID, _ time.Time) error {
			if b.ConferenceLink == nil || *b.ConferenceLink == "" {
				t.Fatal("expected conference link")
			}
			return nil
		},
	}
	mockConf := &conference.MockClient{}
	svc := NewBookingService(repo, mockConf)
	_, err := svc.CreateBooking(context.Background(), uuid.New(), dto.CreateBookingRequest{
		SlotID: slotID, CreateConferenceLink: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateBooking_ConferenceFails(t *testing.T) {
	slotID := uuid.New()
	repo := &fakeBookingRepo{}
	svc := NewBookingService(repo, &conference.MockClient{Err: errors.New("down")})
	_, err := svc.CreateBooking(context.Background(), uuid.New(), dto.CreateBookingRequest{
		SlotID: slotID, CreateConferenceLink: true,
	})
	if err == nil || !errors.Is(err, ErrConferenceFailed) {
		t.Fatalf("got %v", err)
	}
}

func TestCancelBooking_NotFound(t *testing.T) {
	bid := uuid.New()
	repo := &fakeBookingRepo{
		cancelForUser: func(context.Context, uuid.UUID, uuid.UUID) (*model.Booking, error) {
			return nil, repository.ErrCancelBookingNotFound
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	_, err := svc.CancelBooking(context.Background(), uuid.New(), bid)
	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestCancelBooking_NotOwner(t *testing.T) {
	other := uuid.New()
	bid := uuid.New()
	repo := &fakeBookingRepo{
		cancelForUser: func(_ context.Context, u, b uuid.UUID) (*model.Booking, error) {
			if u != other || b != bid {
				t.Fatal()
			}
			return nil, repository.ErrCancelBookingForbidden
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	_, err := svc.CancelBooking(context.Background(), other, bid)
	if !errors.Is(err, ErrCancelNotOwner) {
		t.Fatalf("got %v", err)
	}
}

func TestCancelBooking_IdempotentAlreadyCancelled(t *testing.T) {
	uid := uuid.New()
	bid := uuid.New()
	b := &model.Booking{ID: bid, UserID: uid, Status: "cancelled"}
	repo := &fakeBookingRepo{
		cancelForUser: func(context.Context, uuid.UUID, uuid.UUID) (*model.Booking, error) {
			return b, nil
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	got, err := svc.CancelBooking(context.Background(), uid, bid)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "cancelled" {
		t.Fatalf("got %s", got.Status)
	}
}

func TestCancelBooking_ActiveMarksCancelled(t *testing.T) {
	uid := uuid.New()
	bid := uuid.New()
	cancelled := &model.Booking{ID: bid, UserID: uid, Status: "cancelled"}
	repo := &fakeBookingRepo{
		cancelForUser: func(context.Context, uuid.UUID, uuid.UUID) (*model.Booking, error) {
			return cancelled, nil
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	got, err := svc.CancelBooking(context.Background(), uid, bid)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "cancelled" {
		t.Fatalf("got %s", got.Status)
	}
}

func TestCancelBooking_UnexpectedStatus(t *testing.T) {
	uid := uuid.New()
	bid := uuid.New()
	repo := &fakeBookingRepo{
		cancelForUser: func(context.Context, uuid.UUID, uuid.UUID) (*model.Booking, error) {
			return nil, errors.New("unexpected booking status: pending")
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	_, err := svc.CancelBooking(context.Background(), uid, bid)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListAllBookingsPaged(t *testing.T) {
	want := []*model.Booking{{ID: uuid.New()}}
	repo := &fakeBookingRepo{
		listAllPaged: func(_ context.Context, page, size int) ([]*model.Booking, int64, error) {
			if page != 1 || size != 10 {
				t.Fatalf("pagination")
			}
			return want, 1, nil
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	got, total, err := svc.ListAllBookingsPaged(context.Background(), 1, 10)
	if err != nil || total != 1 || len(got) != 1 {
		t.Fatalf("got %v total %d err %v", got, total, err)
	}
}
