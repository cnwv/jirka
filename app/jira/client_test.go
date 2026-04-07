package jira

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCountJQL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total": 42}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	total, err := client.CountJQL(context.Background(), "project = TEST")
	if err != nil {
		t.Fatal(err)
	}
	if total != 42 {
		t.Errorf("expected 42, got %d", total)
	}
}

func TestSearchJQL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"total": 1,
			"issues": [{
				"key": "TEST-1",
				"fields": {
					"summary": "Test ticket",
					"status": {"name": "Open"},
					"priority": {"name": "High"},
					"issuetype": {"name": "Bug"},
					"created": "2026-01-01T00:00:00.000+0000",
					"updated": "2026-01-02T00:00:00.000+0000"
				}
			}]
		}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	issues, err := client.SearchJQL(context.Background(), "project = TEST")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Key != "TEST-1" {
		t.Errorf("expected key TEST-1, got %q", issues[0].Key)
	}
}

func TestSearchJQL_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorMessages": ["bad query"]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.SearchJQL(context.Background(), "bad jql")
	if err == nil {
		t.Error("expected error for 400 response")
	}
}

func TestSearchJQL_BearerToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total": 0, "issues": []}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "my-secret-token")
	_, _ = client.SearchJQL(context.Background(), "project = TEST")

	expected := "Bearer my-secret-token"
	if gotAuth != expected {
		t.Errorf("expected Authorization %q, got %q", expected, gotAuth)
	}
}
