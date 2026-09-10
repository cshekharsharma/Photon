package httpstub

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ClearAllStubs()
	m.Run()
}

func TestIsStubbingEnabled(t *testing.T) {
	values := []bool{true, false}

	for _, v := range values {
		enableStubs = v
		if IsStubbingEnabled() != v {
			t.Fatalf("expected enableStubs=%v, got %v", enableStubs, !enableStubs)
		}
	}
}

func TestAddStub(t *testing.T) {
	ClearAllStubs()

	entry := &StubEntry{
		Endpoint: "/test",
	}
	id, err := AddStub(entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id == "" {
		t.Fatal("expected a generated id, got empty string")
	}

	st, err := GetHttpStubByID(id)
	if err != nil {
		t.Fatalf("expected stub to be retrievable, got error: %v", err)
	}
	if st.Method != http.MethodGet {
		t.Errorf("expected default method=GET, got %s", st.Method)
	}
	if st.ResponseCode != http.StatusOK {
		t.Errorf("expected default response code=200, got %d", st.ResponseCode)
	}
	if st.ContentType != "application/json" {
		t.Errorf("expected default ContentType=application/json, got %s", st.ContentType)
	}
	if st.Labels == nil {
		t.Error("expected default Labels to be a non-nil map")
	}

	// Add a stub with endpoint empty - should fail
	entry2 := &StubEntry{}
	_, err2 := AddStub(entry2)
	if err2 == nil {
		t.Error("expected error for empty endpoint, got nil")
	}
}

func TestRemoveStub(t *testing.T) {
	ClearAllStubs()
	entry := &StubEntry{Endpoint: "/remove"}
	id, err := AddStub(entry)
	if err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}

	err = RemoveStub(id)
	if err != nil {
		t.Fatalf("expected no error removing existing stub, got %v", err)
	}

	err = RemoveStub("non-existent")
	if err == nil {
		t.Error("expected error removing non-existent stub, got nil")
	}
}

func TestListAllStubs(t *testing.T) {
	ClearAllStubs()
	entry1 := &StubEntry{Endpoint: "/e1"}
	entry2 := &StubEntry{Endpoint: "/e2"}
	if _, err := AddStub(entry1); err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}
	if _, err := AddStub(entry2); err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}

	all := ListAllStubs()
	if len(all) != 2 {
		t.Errorf("expected 2 stubs, got %d", len(all))
	}
}

func TestClearAllStubs(t *testing.T) {
	ClearAllStubs()
	entry := &StubEntry{Endpoint: "/clear"}
	if _, err := AddStub(entry); err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}
	if len(ListAllStubs()) != 1 {
		t.Fatal("expected 1 stub before clearing")
	}

	ClearAllStubs()
	if len(ListAllStubs()) != 0 {
		t.Error("expected 0 stubs after clearing")
	}
}

func TestGetHttpStubByID(t *testing.T) {
	ClearAllStubs()
	entry := &StubEntry{Endpoint: "/get"}
	id, err := AddStub(entry)
	if err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}

	st, err := GetHttpStubByID(id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if st.Endpoint != "/get" {
		t.Errorf("expected endpoint=/get, got %s", st.Endpoint)
	}

	// Empty ID
	_, err2 := GetHttpStubByID("")
	if err2 == nil {
		t.Error("expected error for empty id")
	}

	// Non-existent ID
	_, err3 := GetHttpStubByID("nope")
	if err3 == nil {
		t.Error("expected error for non-existent stub")
	}
}

func TestCallStub(t *testing.T) {
	ClearAllStubs()
	enableStubs = true
	entry := &StubEntry{Endpoint: "/call", PositiveResponse: "Hello"}
	id, err := AddStub(entry)
	if err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}

	ctx := context.Background()
	resp, err := CallStub(ctx, id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response, got nil")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("expected no body close error, got %v", err)
	}
	if string(bodyBytes) != "Hello" {
		t.Errorf("expected body=Hello, got %s", string(bodyBytes))
	}

	entry2 := &StubEntry{
		Endpoint:         "/prob",
		Probability:      0.0, // always negative
		NegativeResponse: "Error",
	}
	id2, err := AddStub(entry2)
	if err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}
	resp2, err2 := CallStub(ctx, id2)
	if err2 != nil {
		t.Fatalf("expected no error, got %v", err2)
	}
	bodyBytes2, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if err := resp2.Body.Close(); err != nil {
		t.Fatalf("expected no body close error, got %v", err)
	}
	if string(bodyBytes2) != "Error" {
		t.Errorf("expected Error body due to Probability=0.0, got %s", string(bodyBytes2))
	}

	entry3 := &StubEntry{
		Endpoint: "/maxhits",
		MaxHits:  1,
	}
	id3, err := AddStub(entry3)
	if err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}
	_, err3 := CallStub(ctx, id3)
	if err3 != nil {
		t.Fatalf("expected no error on first call, got %v", err3)
	}
	_, err4 := CallStub(ctx, id3)
	if err4 == nil {
		t.Error("expected error after max hits exceeded")
	}

	enableStubs = false
	_, err5 := CallStub(ctx, id)
	if err5 != nil {
		t.Errorf("expected no error when stubs disabled, got %v", err5)
	}
	enableStubs = true

	_, err6 := CallStub(ctx, "fakeID")
	if err6 == nil {
		t.Error("expected error for non-existent stub ID")
	}

	entry4 := &StubEntry{
		Endpoint:    "/latency",
		Latency:     200 * time.Millisecond,
		Probability: 1.0,
	}
	id4, err := AddStub(entry4)
	if err != nil {
		t.Fatalf("expected no add stub error, got %v", err)
	}
	ctx2, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	_, err7 := CallStub(ctx2, id4)
	if !errors.Is(err7, context.DeadlineExceeded) {
		t.Errorf("expected context deadline error due to latency, got %v", err7)
	}
}

func TestCallStub_ResponseHeadersAreSet(t *testing.T) {
	ClearAllStubs()
	enableStubs = true

	id, err := AddStub(&StubEntry{
		Endpoint:         "/headers",
		PositiveResponse: "ok",
		ResponseHeaders: map[string]string{
			"X-Test": "yes",
		},
	})
	if err != nil {
		t.Fatalf("unexpected add stub error: %v", err)
	}

	resp, err := CallStub(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected call stub error: %v", err)
	}
	if got := resp.Header.Get("X-Test"); got != "yes" {
		t.Fatalf("expected X-Test=yes, got %s", got)
	}
}

func TestInitStubConfig(t *testing.T) {
	InitStubConfig(false)
	if enableStubs != false {
		t.Errorf("expected enableStub flag to be false, got %v", enableStubs)
	}
}

func TestAddDefaultValues(t *testing.T) {
	entry := &StubEntry{}
	newEntry := addDefaultValues(entry)

	if newEntry.Method != http.MethodGet {
		t.Errorf("expected default method=GET, got %s", newEntry.Method)
	}
	if newEntry.ResponseCode != http.StatusOK {
		t.Errorf("expected default ResponseCode=200, got %d", newEntry.ResponseCode)
	}
	if newEntry.ContentType != "application/json" {
		t.Errorf("expected default ContentType=application/json, got %s", newEntry.ContentType)
	}
	if newEntry.ResponseHeaders == nil {
		t.Error("expected ResponseHeaders not nil")
	}
	if newEntry.Labels == nil {
		t.Error("expected Labels not nil")
	}

	entry2 := &StubEntry{NegativeResponse: "Error"}
	newEntry2 := addDefaultValues(entry2)
	if newEntry2.ErrorResponseCode != http.StatusInternalServerError {
		t.Errorf("expected default ErrorResponseCode=500, got %d", newEntry2.ErrorResponseCode)
	}
}

func TestEnableStubsVariable(t *testing.T) {
	enableStubs = false
	resp, err := CallStub(context.Background(), "nonexistent")
	if err != nil || resp != nil {
		t.Errorf("expected no error and nil resp when stubs disabled, got err=%v resp=%v", err, resp)
	}
	enableStubs = true
}
