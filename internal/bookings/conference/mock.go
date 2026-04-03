package conference

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// MockClient имитирует Conference Service без реального HTTP.
// Поведение при сбоях см. README (раздел про конференции).
type MockClient struct {
	// Err если задано — CreateLink всегда возвращает эту ошибку (имитация недоступности сервиса).
	Err error
}

// NewMockClientFromEnv: при CONFERENCE_MOCK_FAIL=1|true мок всегда падает.
func NewMockClientFromEnv() *MockClient {
	v := os.Getenv("CONFERENCE_MOCK_FAIL")
	if v == "1" || v == "true" {
		return &MockClient{Err: errors.New("conference service unavailable")}
	}
	return &MockClient{}
}

func (m *MockClient) CreateLink(ctx context.Context) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	return fmt.Sprintf("https://conference.mock.example/%s", uuid.NewString()), nil
}
