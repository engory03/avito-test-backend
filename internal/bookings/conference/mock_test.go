package conference

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockClient_CreateLink_Success(t *testing.T) {
	m := &MockClient{}
	url, err := m.CreateLink(context.Background())
	if err != nil || url == "" {
		t.Fatalf("%s %v", url, err)
	}
}

func TestMockClient_CreateLink_Error(t *testing.T) {
	m := &MockClient{Err: errors.New("fail")}
	_, err := m.CreateLink(context.Background())
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestMockClient_CreateLink_ContextCancelled(t *testing.T) {
	m := &MockClient{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.CreateLink(ctx)
	if err == nil {
		t.Fatal("expected ctx err")
	}
}

func TestNewMockClientFromEnv_Default(t *testing.T) {
	t.Setenv("CONFERENCE_MOCK_FAIL", "")
	c := NewMockClientFromEnv()
	if c.Err != nil {
		t.Fatal(c.Err)
	}
}

func TestNewMockClientFromEnv_Fail(t *testing.T) {
	t.Setenv("CONFERENCE_MOCK_FAIL", "1")
	c := NewMockClientFromEnv()
	if c.Err == nil {
		t.Fatal("expected Err")
	}
}

func TestNewMockClientFromEnv_FailTrue(t *testing.T) {
	t.Setenv("CONFERENCE_MOCK_FAIL", "true")
	c := NewMockClientFromEnv()
	if c.Err == nil {
		t.Fatal("expected Err")
	}
}

// Exercise select default branch with non-cancelled context.
func TestMockClient_CreateLink_NonBlocking(t *testing.T) {
	m := &MockClient{}
	ctx := context.Background()
	done := make(chan string, 1)
	go func() {
		u, err := m.CreateLink(ctx)
		if err != nil {
			done <- ""
			return
		}
		done <- u
	}()
	select {
	case u := <-done:
		if u == "" {
			t.Fatal("link")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
