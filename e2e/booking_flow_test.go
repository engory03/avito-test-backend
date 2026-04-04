package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-engory03/internal/auth/dummy"
)

func baseURL(t *testing.T) string {
	t.Helper()
	u := strings.TrimSpace(os.Getenv("E2E_BASE_URL"))
	if u == "" {
		t.Skip("set E2E_BASE_URL (e.g. http://localhost:8080) to run E2E tests")
	}
	return strings.TrimRight(u, "/")
}

func TestRoomScheduleBookingFlow(t *testing.T) {
	base := baseURL(t)
	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	adminTok := dummyLogin(t, ctx, client, base, "admin")
	userTok := dummyLogin(t, ctx, client, base, "user")

	roomName := fmt.Sprintf("e2e-room-%d", time.Now().UnixNano())
	roomID := createRoom(t, ctx, client, base, adminTok, roomName)
	createSchedule(t, ctx, client, base, adminTok, roomID)

	slotID := pickFutureSlotID(t, ctx, client, base, userTok, roomID)
	booking := createBooking(t, ctx, client, base, userTok, slotID)

	if booking.Status != "active" {
		t.Fatalf("booking status: got %q want active", booking.Status)
	}
	if booking.SlotID != slotID {
		t.Fatalf("booking.slotId: got %s want %s", booking.SlotID, slotID)
	}
	if booking.UserID != dummy.DummyUserID {
		t.Fatalf("booking.userId: got %s want dummy user %s", booking.UserID, dummy.DummyUserID)
	}
}

// --- HTTP helpers ---

type tokenResponse struct {
	Token string `json:"token"`
}

type apiErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func dummyLogin(t *testing.T, ctx context.Context, client *http.Client, base, role string) string {
	t.Helper()
	body := map[string]string{"role": role}
	var out tokenResponse
	status, raw := doJSON(t, ctx, client, http.MethodPost, base+"/api/dummyLogin", "", body, &out)
	if status != http.StatusOK {
		t.Fatalf("dummyLogin %s: status %d %s", role, status, apiErrDetail(raw))
	}
	if out.Token == "" {
		t.Fatalf("dummyLogin %s: empty token", role)
	}
	return out.Token
}

type roomEnvelope struct {
	Room struct {
		ID uuid.UUID `json:"id"`
	} `json:"room"`
}

func createRoom(t *testing.T, ctx context.Context, client *http.Client, base, adminToken, name string) uuid.UUID {
	t.Helper()
	reqBody := map[string]any{"name": name}
	var out roomEnvelope
	status, raw := doJSON(t, ctx, client, http.MethodPost, base+"/api/rooms/create", adminToken, reqBody, &out)
	if status != http.StatusCreated {
		t.Fatalf("create room: status %d %s", status, apiErrDetail(raw))
	}
	if out.Room.ID == uuid.Nil {
		t.Fatal("create room: missing room.id")
	}
	return out.Room.ID
}

func createSchedule(t *testing.T, ctx context.Context, client *http.Client, base, adminToken string, roomID uuid.UUID) {
	t.Helper()
	url := fmt.Sprintf("%s/api/rooms/%s/schedule/create", base, roomID.String())
	reqBody := map[string]any{
		"roomId":     roomID.String(),
		"daysOfWeek": []int{1, 2, 3, 4, 5},
		"startTime":  "09:00",
		"endTime":    "18:00",
	}
	status, raw := doJSON(t, ctx, client, http.MethodPost, url, adminToken, reqBody, nil)
	if status != http.StatusCreated {
		t.Fatalf("create schedule: status %d %s", status, apiErrDetail(raw))
	}
}

type slotJSON struct {
	ID    string `json:"id"`
	Start string `json:"start"`
}

type slotsEnvelope struct {
	Slots []slotJSON `json:"slots"`
}

func listSlots(t *testing.T, ctx context.Context, client *http.Client, base, token string, roomID uuid.UUID, dateYYYYMMDD string) []slotJSON {
	t.Helper()
	url := fmt.Sprintf("%s/api/rooms/%s/slots/list?date=%s", base, roomID.String(), dateYYYYMMDD)
	var out slotsEnvelope
	status, raw := doJSON(t, ctx, client, http.MethodGet, url, token, nil, &out)
	if status != http.StatusOK {
		t.Fatalf("list slots date=%s: status %d %s", dateYYYYMMDD, status, apiErrDetail(raw))
	}
	return out.Slots
}

func apiWeekdayUTC(day time.Time) int {
	wd := day.Weekday()
	if wd == time.Sunday {
		return 7
	}
	return int(wd)
}

func pickFutureSlotID(t *testing.T, ctx context.Context, client *http.Client, base, token string, roomID uuid.UUID) uuid.UUID {
	t.Helper()
	now := time.Now().UTC()
	for i := 1; i <= 21; i++ {
		candidate := now.AddDate(0, 0, i)
		y, m, d := candidate.Date()
		dayUTC := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		w := apiWeekdayUTC(dayUTC)
		if w < 1 || w > 5 {
			continue
		}
		dateStr := dayUTC.Format("2006-01-02")
		slots := listSlots(t, ctx, client, base, token, roomID, dateStr)
		for _, s := range slots {
			start, err := time.Parse(time.RFC3339Nano, s.Start)
			if err != nil {
				start, err = time.Parse(time.RFC3339, s.Start)
			}
			if err != nil {
				t.Fatalf("parse slot start %q: %v", s.Start, err)
			}
			if start.After(now) {
				id, err := uuid.Parse(s.ID)
				if err != nil {
					t.Fatalf("parse slot id %q: %v", s.ID, err)
				}
				return id
			}
		}
	}
	t.Fatal("no future weekday slot found in 3 weeks; check schedule / server time")
	return uuid.Nil
}

type bookingJSON struct {
	ID     uuid.UUID `json:"id"`
	SlotID uuid.UUID `json:"slotId"`
	UserID uuid.UUID `json:"userId"`
	Status string    `json:"status"`
}

type bookingEnvelope struct {
	Booking bookingJSON `json:"booking"`
}

func createBooking(t *testing.T, ctx context.Context, client *http.Client, base, userToken string, slotID uuid.UUID) bookingJSON {
	t.Helper()
	reqBody := map[string]any{"slotId": slotID.String()}
	var out bookingEnvelope
	status, raw := doJSON(t, ctx, client, http.MethodPost, base+"/api/bookings/create", userToken, reqBody, &out)
	if status != http.StatusCreated {
		t.Fatalf("create booking: status %d %s", status, apiErrDetail(raw))
	}
	return out.Booking
}

func doJSON(t *testing.T, ctx context.Context, client *http.Client, method, url, bearer string, reqBody any, respBody any) (int, []byte) {
	t.Helper()
	var req *http.Request
	var err error
	if method == http.MethodGet || reqBody == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		b, mErr := json.Marshal(reqBody)
		if mErr != nil {
			t.Fatalf("marshal body: %v", mErr)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, url, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if respBody != nil && len(raw) > 0 {
		if uErr := json.Unmarshal(raw, respBody); uErr != nil {
			if res.StatusCode < 300 {
				t.Fatalf("decode JSON (%s %s): %v body=%s", method, url, uErr, raw)
			}
		}
	}
	return res.StatusCode, raw
}

func apiErrDetail(raw []byte) string {
	var ae apiErrorBody
	if json.Unmarshal(raw, &ae) == nil && strings.TrimSpace(ae.Error.Message) != "" {
		return fmt.Sprintf("%s: %s", ae.Error.Code, ae.Error.Message)
	}
	return string(raw)
}
