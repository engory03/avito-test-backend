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
	slotByID          func(context.Context, uuid.UUID) (*roomsModel.Slot, error)
	createBooking     func(context.Context, *model.Booking, uuid.UUID) error
	getByID           func(context.Context, uuid.UUID) (*model.Booking, error)
	markCancelled     func(context.Context, uuid.UUID) error
	listMy            func(context.Context, uuid.UUID, time.Time) ([]*model.Booking, error)
	listAllPaged      func(context.Context, int, int) ([]*model.Booking, int64, error)
	lastGetByIDResult *model.Booking
}

func (f *fakeBookingRepo) ListByUserWithFutureSlots(ctx context.Context, userID uuid.UUID, nowUTC time.Time) ([]*model.Booking, error) {
	if f.listMy != nil {
		return f.listMy(ctx, userID, nowUTC)
	}
	return nil, nil
}

func (f *fakeBookingRepo) GetSlotByID(ctx context.Context, slotID uuid.UUID) (*roomsModel.Slot, error) {
	if f.slotByID != nil {
		return f.slotByID(ctx, slotID)
	}
	return nil, nil
}

func (f *fakeBookingRepo) CreateActiveBookingIfSlotFree(ctx context.Context, booking *model.Booking, slotID uuid.UUID) error {
	if f.createBooking != nil {
		return f.createBooking(ctx, booking, slotID)
	}
	return nil
}

func (f *fakeBookingRepo) GetByID(ctx context.Context, bookingID uuid.UUID) (*model.Booking, error) {
	if f.getByID != nil {
		b, err := f.getByID(ctx, bookingID)
		f.lastGetByIDResult = b
		return b, err
	}
	return nil, nil
}

func (f *fakeBookingRepo) MarkActiveBookingCancelled(ctx context.Context, bookingID uuid.UUID) error {
	if f.markCancelled != nil {
		return f.markCancelled(ctx, bookingID)
	}
	return nil
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
		slotByID: func(_ context.Context, id uuid.UUID) (*roomsModel.Slot, error) {
			if id != slotID {
				t.Fatalf("slot id")
			}
			return nil, nil
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
		slotByID: func(_ context.Context, _ uuid.UUID) (*roomsModel.Slot, error) {
			return &roomsModel.Slot{
				ID:      slotID,
				StartAt: time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
			}, nil
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
	future := time.Now().UTC().Add(24 * time.Hour)
	repo := &fakeBookingRepo{
		slotByID: func(_ context.Context, _ uuid.UUID) (*roomsModel.Slot, error) {
			return &roomsModel.Slot{ID: slotID, StartAt: future}, nil
		},
		createBooking: func(_ context.Context, b *model.Booking, sid uuid.UUID) error {
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
	future := time.Now().UTC().Add(48 * time.Hour)
	repo := &fakeBookingRepo{
		slotByID: func(_ context.Context, _ uuid.UUID) (*roomsModel.Slot, error) {
			return &roomsModel.Slot{ID: slotID, StartAt: future}, nil
		},
		createBooking: func(context.Context, *model.Booking, uuid.UUID) error {
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
	future := time.Now().UTC().Add(72 * time.Hour)
	repo := &fakeBookingRepo{
		slotByID: func(_ context.Context, _ uuid.UUID) (*roomsModel.Slot, error) {
			return &roomsModel.Slot{ID: slotID, StartAt: future}, nil
		},
		createBooking: func(_ context.Context, b *model.Booking, _ uuid.UUID) error {
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
	future := time.Now().UTC().Add(96 * time.Hour)
	repo := &fakeBookingRepo{
		slotByID: func(_ context.Context, _ uuid.UUID) (*roomsModel.Slot, error) {
			return &roomsModel.Slot{ID: slotID, StartAt: future}, nil
		},
	}
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
		getByID: func(_ context.Context, id uuid.UUID) (*model.Booking, error) {
			if id != bid {
				t.Fatal()
			}
			return nil, nil
		},
	}
	svc := NewBookingService(repo, conference.NewMockClientFromEnv())
	_, err := svc.CancelBooking(context.Background(), uuid.New(), bid)
	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestCancelBooking_NotOwner(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	bid := uuid.New()
	repo := &fakeBookingRepo{
		getByID: func(context.Context, uuid.UUID) (*model.Booking, error) {
			return &model.Booking{ID: bid, UserID: owner, Status: "active"}, nil
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
		getByID: func(context.Context, uuid.UUID) (*model.Booking, error) {
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
	call := 0
	repo := &fakeBookingRepo{
		getByID: func(context.Context, uuid.UUID) (*model.Booking, error) {
			call++
			if call == 1 {
				return &model.Booking{ID: bid, UserID: uid, Status: "active"}, nil
			}
			return cancelled, nil
		},
		markCancelled: func(_ context.Context, id uuid.UUID) error {
			if id != bid {
				t.Fatal()
			}
			return nil
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
		getByID: func(context.Context, uuid.UUID) (*model.Booking, error) {
			return &model.Booking{ID: bid, UserID: uid, Status: "pending"}, nil
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
