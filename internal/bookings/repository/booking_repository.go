package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/bookings/model"
	roomsModel "github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrActiveBookingExists — слот уже имеет активную бронь (проверка или UNIQUE partial index).
	ErrActiveBookingExists = errors.New("active booking exists for slot")
	// ErrBookingSlotNotFound — в транзакции слот не найден (после FOR UPDATE).
	ErrBookingSlotNotFound = errors.New("slot not found")
	// ErrBookingSlotInPast — start_at слота раньше nowUTC.
	ErrBookingSlotInPast = errors.New("slot is in the past")
	// ErrCancelBookingNotFound — бронь не найдена (отмена).
	ErrCancelBookingNotFound = errors.New("booking not found for cancel")
	// ErrCancelBookingForbidden — не владелец брони.
	ErrCancelBookingForbidden = errors.New("forbidden: not booking owner")
)

type BookingRepository interface {
	ListByUserWithFutureSlots(ctx context.Context, userID uuid.UUID, nowUTC time.Time) ([]*model.Booking, error)
	GetSlotByID(ctx context.Context, slotID uuid.UUID) (*roomsModel.Slot, error)
	CreateActiveBookingWithSlotLock(ctx context.Context, booking *model.Booking, slotID uuid.UUID, nowUTC time.Time) error
	CancelBookingForUser(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error)
	GetByID(ctx context.Context, bookingID uuid.UUID) (*model.Booking, error)
	ListAllPaged(ctx context.Context, page, pageSize int) (bookings []*model.Booking, total int64, err error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func isPostgresUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func (r *bookingRepository) ListByUserWithFutureSlots(ctx context.Context, userID uuid.UUID, nowUTC time.Time) ([]*model.Booking, error) {
	var rows []*model.Booking
	err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Joins("INNER JOIN slots ON slots.id = bookings.slot_id").
		Where("bookings.user_id = ? AND slots.start_at >= ?", userID, nowUTC).
		Order("slots.start_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *bookingRepository) GetSlotByID(ctx context.Context, slotID uuid.UUID) (*roomsModel.Slot, error) {
	var s roomsModel.Slot
	err := r.db.WithContext(ctx).First(&s, "id = ?", slotID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CreateActiveBookingWithSlotLock: REPEATABLE READ, SELECT slot FOR UPDATE, проверки, INSERT active.
func (r *bookingRepository) CreateActiveBookingWithSlotLock(ctx context.Context, booking *model.Booking, slotID uuid.UUID, nowUTC time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var slot roomsModel.Slot
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", slotID).
			First(&slot).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBookingSlotNotFound
		}
		if err != nil {
			return err
		}
		if slot.StartAt.UTC().Before(nowUTC) {
			return ErrBookingSlotInPast
		}

		var n int64
		if err := tx.Model(&model.Booking{}).
			Where("slot_id = ? AND status = ?", slotID, "active").
			Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return ErrActiveBookingExists
		}

		if err := tx.Create(booking).Error; err != nil {
			if isPostgresUniqueViolation(err) {
				return ErrActiveBookingExists
			}
			return err
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
}

// CancelBookingForUser: READ COMMITTED, SELECT booking FOR UPDATE, идемпотентная отмена.
func (r *bookingRepository) CancelBookingForUser(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error) {
	var out *model.Booking
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var b model.Booking
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", bookingID).
			First(&b).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCancelBookingNotFound
		}
		if err != nil {
			return err
		}
		if b.UserID != userID {
			return ErrCancelBookingForbidden
		}
		if b.Status == "cancelled" {
			out = &b
			return nil
		}
		if b.Status != "active" {
			return fmt.Errorf("unexpected booking status: %s", b.Status)
		}

		now := time.Now().UTC()
		res := tx.Model(&model.Booking{}).
			Where("id = ? AND status = ?", bookingID, "active").
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if err := tx.Where("id = ?", bookingID).First(&b).Error; err != nil {
			return err
		}
		out = &b
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *bookingRepository) GetByID(ctx context.Context, bookingID uuid.UUID) (*model.Booking, error) {
	var b model.Booking
	err := r.db.WithContext(ctx).First(&b, "id = ?", bookingID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepository) ListAllPaged(ctx context.Context, page, pageSize int) ([]*model.Booking, int64, error) {
	var total int64
	base := r.db.WithContext(ctx).Model(&model.Booking{})
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []*model.Booking
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
