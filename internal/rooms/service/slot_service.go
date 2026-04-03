package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/model"
	"github.com/internships-backend/test-backend-engory03/internal/rooms/repository"
)

var ErrInvalidSlotListDate = errors.New("invalid date parameter")

const slotStep = 30 * time.Minute

type SlotService interface {
	ListAvailableSlots(roomID uuid.UUID, dateStr string) ([]model.Slot, error)
}

type slotService struct {
	roomRepo     repository.RoomRepository
	scheduleRepo repository.ScheduleRepository
	slotRepo     repository.SlotRepository
}

func NewSlotService(
	roomRepo repository.RoomRepository,
	scheduleRepo repository.ScheduleRepository,
	slotRepo repository.SlotRepository,
) SlotService {
	return &slotService{
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
		slotRepo:     slotRepo,
	}
}

func (s *slotService) ListAvailableSlots(roomID uuid.UUID, dateStr string) ([]model.Slot, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return nil, ErrInvalidSlotListDate
	}
	dayUTC, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%w: use ISO date YYYY-MM-DD", ErrInvalidSlotListDate)
	}

	ok, err := s.roomRepo.ExistsByID(roomID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrRoomNotFound
	}

	sched, err := s.scheduleRepo.GetByRoomID(roomID)
	if err != nil {
		return nil, err
	}
	if sched == nil {
		return nil, nil
	}

	if !containsInt(sched.DaysOfWeek, apiWeekday(dayUTC)) {
		return nil, nil
	}

	startClock, err := parseHHMM(sched.StartTime)
	if err != nil {
		return nil, err
	}
	endClock, err := parseHHMM(sched.EndTime)
	if err != nil {
		return nil, err
	}
	y, m, d := dayUTC.Date()
	windowStart := time.Date(y, m, d, startClock.h, startClock.m, 0, 0, time.UTC)
	windowEnd := time.Date(y, m, d, endClock.h, endClock.m, 0, 0, time.UTC)
	if !windowEnd.After(windowStart) {
		return nil, nil
	}

	var candidates []model.Slot
	for t := windowStart; t.Add(slotStep).Compare(windowEnd) <= 0; t = t.Add(slotStep) {
		end := t.Add(slotStep)
		id := deterministicSlotID(roomID, t)
		candidates = append(candidates, model.Slot{
			ID:      id,
			RoomID:  roomID,
			StartAt: t,
			EndAt:   end,
		})
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	for i := range candidates {
		if err := s.slotRepo.UpsertGeneratedSlot(&candidates[i]); err != nil {
			return nil, err
		}
	}

	ids := make([]uuid.UUID, len(candidates))
	for i := range candidates {
		ids[i] = candidates[i].ID
	}
	booked, err := s.slotRepo.ListActiveBookedSlotIDs(ids)
	if err != nil {
		return nil, err
	}

	free := make([]model.Slot, 0, len(candidates))
	for _, sl := range candidates {
		if _, taken := booked[sl.ID]; !taken {
			free = append(free, sl)
		}
	}
	return free, nil
}

func apiWeekday(t time.Time) int {
	wd := t.Weekday()
	if wd == time.Sunday {
		return 7
	}
	return int(wd)
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

type clock struct{ h, m int }

func parseHHMM(s string) (clock, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return clock{}, fmt.Errorf("invalid time %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return clock{}, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return clock{}, err
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return clock{}, fmt.Errorf("invalid time %q", s)
	}
	return clock{h: h, m: m}, nil
}

func deterministicSlotID(roomID uuid.UUID, startUTC time.Time) uuid.UUID {
	key := append(append(roomID[:], '|'), []byte(startUTC.UTC().Format("2006-01-02T15:04:05Z"))...)
	return uuid.NewSHA1(uuid.NameSpaceURL, key)
}
