package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/dto"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/service"
)

type stubBookingSvc struct {
	listMy  func(uuid.UUID) ([]*model.Booking, error)
	create  func(context.Context, uuid.UUID, dto.CreateBookingRequest) (*model.Booking, error)
	cancel  func(uuid.UUID, uuid.UUID) (*model.Booking, error)
	listAll func(int, int) ([]*model.Booking, int64, error)
}

func (s *stubBookingSvc) ListMyFutureBookings(userID uuid.UUID) ([]*model.Booking, error) {
	if s.listMy != nil {
		return s.listMy(userID)
	}
	return nil, nil
}

func (s *stubBookingSvc) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.CreateBookingRequest) (*model.Booking, error) {
	if s.create != nil {
		return s.create(ctx, userID, req)
	}
	return nil, errors.New("noop")
}

func (s *stubBookingSvc) CancelBooking(userID, bookingID uuid.UUID) (*model.Booking, error) {
	if s.cancel != nil {
		return s.cancel(userID, bookingID)
	}
	return nil, errors.New("noop")
}

func (s *stubBookingSvc) ListAllBookingsPaged(page, pageSize int) ([]*model.Booking, int64, error) {
	if s.listAll != nil {
		return s.listAll(page, pageSize)
	}
	return nil, 0, nil
}

func newCtxWithUser(t *testing.T, uid uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uid)
	return c, w
}

func TestBookingHandler_MyBookings_MissingUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h := NewBookingHandler(&stubBookingSvc{})
	h.MyBookings(c)
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
}

func TestBookingHandler_MyBookings_OK(t *testing.T) {
	uid := uuid.New()
	b := &model.Booking{ID: uuid.New(), UserID: uid, Status: "active", CreatedAt: time.Now().UTC()}
	c, w := newCtxWithUser(t, uid)
	h := NewBookingHandler(&stubBookingSvc{
		listMy: func(u uuid.UUID) ([]*model.Booking, error) {
			if u != uid {
				t.Fatal()
			}
			return []*model.Booking{b}, nil
		},
	})
	h.MyBookings(c)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}

func TestBookingHandler_CreateBooking_Errors(t *testing.T) {
	uid := uuid.New()
	slotID := uuid.New()
	body := func() *bytes.Buffer {
		b, _ := json.Marshal(dto.CreateBookingRequest{SlotID: slotID})
		return bytes.NewBuffer(b)
	}
	tests := []struct {
		name string
		err  error
		code int
	}{
		{"notfound", service.ErrSlotNotFound, 404},
		{"past", service.ErrSlotInPast, 400},
		{"booked", service.ErrSlotAlreadyBooked, 409},
		{"conf", service.ErrConferenceFailed, 500},
		{"other", errors.New("x"), 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newCtxWithUser(t, uid)
			c.Request = httptest.NewRequest(http.MethodPost, "/", body())
			c.Request.Header.Set("Content-Type", "application/json")
			h := NewBookingHandler(&stubBookingSvc{
				create: func(_ context.Context, u uuid.UUID, r dto.CreateBookingRequest) (*model.Booking, error) {
					if u != uid || r.SlotID != slotID {
						t.Fatal()
					}
					return nil, tc.err
				},
			})
			h.CreateBooking(c)
			if w.Code != tc.code {
				t.Fatalf("want %d got %d", tc.code, w.Code)
			}
		})
	}
}

func TestBookingHandler_CreateBooking_Created(t *testing.T) {
	uid := uuid.New()
	slotID := uuid.New()
	b := &model.Booking{ID: uuid.New(), SlotID: slotID, UserID: uid, Status: "active", CreatedAt: time.Now().UTC()}
	payload, _ := json.Marshal(dto.CreateBookingRequest{SlotID: slotID})
	c, w := newCtxWithUser(t, uid)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	h := NewBookingHandler(&stubBookingSvc{
		create: func(_ context.Context, u uuid.UUID, r dto.CreateBookingRequest) (*model.Booking, error) {
			return b, nil
		},
	})
	h.CreateBooking(c)
	if w.Code != 201 {
		t.Fatal(w.Code)
	}
}

func TestBookingHandler_CancelBooking_Errors(t *testing.T) {
	uid := uuid.New()
	bid := uuid.New()
	for _, tc := range []struct {
		path string
		code int
		err  error
	}{
		{"/" + bid.String() + "/cancel", 404, service.ErrBookingNotFound},
		{"/" + bid.String() + "/cancel", 403, service.ErrCancelNotOwner},
		{"/" + bid.String() + "/cancel", 500, errors.New("x")},
	} {
		c, w := newCtxWithUser(t, uid)
		c.Request = httptest.NewRequest(http.MethodPost, tc.path, http.NoBody)
		c.Params = gin.Params{{Key: "bookingId", Value: bid.String()}}
		h := NewBookingHandler(&stubBookingSvc{
			cancel: func(u, b uuid.UUID) (*model.Booking, error) {
				return nil, tc.err
			},
		})
		h.CancelBooking(c)
		if w.Code != tc.code {
			t.Fatalf("path %s want %d got %d", tc.path, tc.code, w.Code)
		}
	}
}

func TestBookingHandler_CancelBooking_InvalidUUID(t *testing.T) {
	c, w := newCtxWithUser(t, uuid.New())
	c.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	c.Params = gin.Params{{Key: "bookingId", Value: "not-uuid"}}
	NewBookingHandler(&stubBookingSvc{}).CancelBooking(c)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}

func TestBookingHandler_ListAllBookings(t *testing.T) {
	c, w := newCtxWithUser(t, uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/list?page=1&pageSize=10", http.NoBody)
	h := NewBookingHandler(&stubBookingSvc{
		listAll: func(page, ps int) ([]*model.Booking, int64, error) {
			if page != 1 || ps != 10 {
				t.Fatal()
			}
			return []*model.Booking{{ID: uuid.New()}}, 1, nil
		},
	})
	h.ListAllBookings(c)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}

func TestBookingHandler_ListAllBookings_BadPagination(t *testing.T) {
	c, w := newCtxWithUser(t, uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/list?page=0", http.NoBody)
	NewBookingHandler(&stubBookingSvc{}).ListAllBookings(c)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
