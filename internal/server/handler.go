package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vrnvu/temp/pkg/quiz"
)

const (
	headerContentType    = "Content-Type"
	valueContentTypeJSON = "application/json"
	headerXRequestID     = "X-Request-ID"
)

type xRequestIDHeader string

const xRequestIDHeaderKey xRequestIDHeader = headerXRequestID

type Handler struct {
	Slog *slog.Logger
	Mux  *http.ServeMux
	db   *InMemoryDB
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Mux.ServeHTTP(w, r)
}

type Config struct {
	Slog               *slog.Logger
	RequestIDGenerator func() string
}

func FromConfig(c *Config) (*Handler, error) {
	db, err := NewInMemoryDB()
	if err != nil {
		return nil, err
	}

	h := &Handler{Slog: c.Slog, Mux: http.NewServeMux(), db: db}

	h.Mux.HandleFunc("GET /health", withBaseMiddleware(h.Slog, c.RequestIDGenerator, health))
	h.Mux.HandleFunc("GET /quiz", withBaseMiddleware(h.Slog, c.RequestIDGenerator, h.getQuiz))
	h.Mux.HandleFunc("GET /quiz/{user}", withBaseMiddleware(h.Slog, c.RequestIDGenerator, h.getQuizResults))
	h.Mux.HandleFunc("PUT /quiz/{user}", withBaseMiddleware(h.Slog, c.RequestIDGenerator, h.putQuizAnswers))
	h.Mux.HandleFunc("PUT /users/{user}", withBaseMiddleware(h.Slog, c.RequestIDGenerator, h.putNewUser))
	h.Mux.HandleFunc("GET /statistics/{user}", withBaseMiddleware(h.Slog, c.RequestIDGenerator, h.getStatistics))
	return h, nil
}

func health(_ http.ResponseWriter, _ *http.Request) {}

func (h *Handler) getQuiz(w http.ResponseWriter, r *http.Request) {
	questions, err := h.db.GetQuestions(r.Context())
	if err != nil {
		h.logError(r, http.StatusText(http.StatusInternalServerError), err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, r, questions)
}

func (h *Handler) putQuizAnswers(w http.ResponseWriter, r *http.Request) {
	user, err := fromPathUser(r)
	if err != nil {
		h.logError(r, http.StatusText(http.StatusBadRequest), err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	quizAnswer := quiz.QuizAnswer{}
	if err := json.NewDecoder(r.Body).Decode(&quizAnswer); err != nil {
		h.logError(r, http.StatusText(http.StatusBadRequest), err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.InsertQuizAnswer(r.Context(), user, quizAnswer); err != nil {
		switch err {
		case ErrUserNotFound:
			h.logError(r, http.StatusText(http.StatusBadRequest), err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			h.logError(r, http.StatusText(http.StatusInternalServerError), err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) getQuizResults(w http.ResponseWriter, r *http.Request) {
	user, err := fromPathUser(r)
	if err != nil {
		h.logError(r, http.StatusText(http.StatusBadRequest), err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := h.db.GetResults(r.Context(), user)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			h.logError(r, http.StatusText(http.StatusBadRequest), err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			h.logError(r, http.StatusText(http.StatusInternalServerError), err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	h.writeJSON(w, r, results)
}

func (h *Handler) putNewUser(w http.ResponseWriter, r *http.Request) {
	user, err := fromPathUser(r)
	if err != nil {
		h.logError(r, http.StatusText(http.StatusBadRequest), err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.InsertUser(r.Context(), user); err != nil {
		switch err {
		case ErrUserAlreadyExists:
			h.logError(r, http.StatusText(http.StatusBadRequest), err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			h.logError(r, http.StatusText(http.StatusInternalServerError), err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) getStatistics(w http.ResponseWriter, r *http.Request) {
	user, err := fromPathUser(r)
	if err != nil {
		h.logError(r, http.StatusText(http.StatusBadRequest), err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statistics, err := h.db.GetStatistics(r.Context(), user)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			h.logError(r, http.StatusText(http.StatusBadRequest), err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case ErrNotEnoughUsersForStatistics:
			h.logError(r, http.StatusText(http.StatusBadRequest), err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			h.logError(r, http.StatusText(http.StatusInternalServerError), err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	h.writeJSON(w, r, statistics)
}

func (h *Handler) writeJSON(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set(headerContentType, valueContentTypeJSON)
	w.Header().Set(headerXRequestID, fromContext(r, xRequestIDHeaderKey))

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logError(r, http.StatusText(http.StatusInternalServerError), err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) logError(r *http.Request, message string, err error) {
	h.Slog.Error(message, "error", err, "method", r.Method, "path", r.URL.Path, headerXRequestID, fromContext(r, xRequestIDHeaderKey))
}

func fromContext(r *http.Request, key any) string {
	return r.Context().Value(key).(string)
}

func withRequestID(requestIDGenerator func() string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := requestIDGenerator()
		ctx := context.WithValue(r.Context(), xRequestIDHeaderKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func withLoggingMethod(slog *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("", "method", r.Method, "path", r.URL.Path, headerXRequestID, fromContext(r, xRequestIDHeaderKey))
		next.ServeHTTP(w, r)
	}
}

func withBaseMiddleware(slog *slog.Logger, requestIDGenerator func() string, next http.HandlerFunc) http.HandlerFunc {
	return withRequestID(requestIDGenerator, withLoggingMethod(slog, next))
}

func assertHeaderValueIs(r *http.Request, header string, value string) error {
	if r.Header.Get(header) != value {
		return fmt.Errorf("invalid header `%s` value: got `%s`, use `%s`", header, r.Header.Get(header), value)
	}
	return nil
}

func fromPathUser(r *http.Request) (string, error) {
	rawUser := r.PathValue("user")
	if rawUser == "" {
		return "", fmt.Errorf("invalid user: `%s`", rawUser)
	}
	return rawUser, nil
}
