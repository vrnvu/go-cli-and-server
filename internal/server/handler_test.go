package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/vrnvu/temp/pkg/quiz"
)

func testHandler(t *testing.T) *Handler {
	handler, err := FromConfig(&Config{
		Slog: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		RequestIDGenerator: func() string {
			return "123"
		},
	})
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}
	return handler
}

func TestHandlerHealth(t *testing.T) {
	t.Parallel()
	handler := testHandler(t)
	r, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandlerQuiz(t *testing.T) {
	t.Parallel()
	handler := testHandler(t)

	r, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/quiz", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var body []quiz.Question
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if len(body) == 0 {
		t.Fatalf("expected body to be bigger than zero, got %d", len(body))
	}

}

func TestHandlerQuizResults(t *testing.T) {
	t.Parallel()
	handler := testHandler(t)
	r, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/quiz/user", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var body quiz.QuizResults
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if body.Correct != 0 {
		t.Fatalf("expected correct to be zero, got %d", body.Correct)
	}

	if body.Total != 0 {
		t.Fatalf("expected total to be zero, got %d", body.Total)
	}
}

func TestHandlerPutQuizAnswers(t *testing.T) {
	t.Parallel()
	handler := testHandler(t)
	r, err := http.NewRequestWithContext(context.Background(), http.MethodPut, "/quiz/user", strings.NewReader(`{"1": "a"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	r, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "/quiz/user", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w = httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var body quiz.QuizResults
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if body.Correct != 0 {
		t.Fatalf("expected correct to be zero, got %d", body.Correct)
	}

	if body.Total != 1 {
		t.Fatalf("expected total to be one, got %d", body.Total)
	}
}

func TestHandlerPutNewUser(t *testing.T) {
	handler := testHandler(t)
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "newUser", statusCode: http.StatusOK},
		{name: "user", statusCode: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/users/%s", tt.name)
			r, err := http.NewRequestWithContext(context.Background(), http.MethodPut, url, nil)

			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			w := httptest.NewRecorder()
			handler.Mux.ServeHTTP(w, r)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status code %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestHandlerGetStatistics(t *testing.T) {
	t.Parallel()
	handler := testHandler(t)

	r, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/statistics/user", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	r, err = http.NewRequestWithContext(context.Background(), http.MethodPut, "/users/user1", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w = httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	r, err = http.NewRequestWithContext(context.Background(), http.MethodPut, "/users/user2", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w = httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	r, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "/statistics/user", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w = httptest.NewRecorder()
	handler.Mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expected := quiz.StatisticsResults{
		Correct:    0,
		Total:      0,
		AvgCorrect: 0,
		AvgTotal:   0,
	}

	var body quiz.StatisticsResults
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if body != expected {
		t.Fatalf("expected body %v, got %v", expected, body)
	}
}
