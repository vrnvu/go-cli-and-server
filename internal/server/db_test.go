package server

import (
	"context"
	"testing"

	"github.com/vrnvu/temp/pkg/quiz"
)

func TestInMemoryDB(t *testing.T) {
	_, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Error creating in-memory database: %v", err)
	}
}

func TestGetUserID(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Error creating in-memory database: %v", err)
	}

	tests := []struct {
		name  string
		user  string
		isErr bool
		want  uint64
	}{
		{name: "get success", user: "user", isErr: false, want: 0},
		{name: "get error", user: "", isErr: true, want: 123},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID, err := db.getUserID(test.user)
			if test.isErr {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}

			if userID != test.want {
				t.Fatalf("Expected user ID %d, got %d", test.want, userID)
			}
		})
	}
}

func TestGetQuestions(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Error creating in-memory database: %v", err)
	}

	questions, err := db.GetQuestions(context.Background())
	if err != nil {
		t.Fatalf("Error getting questions: %v", err)
	}

	if len(questions) != 2 {
		t.Fatalf("Expected 2 questions, got %d", len(questions))
	}
}

func TestInsertQuizAnswerAndGetResults(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Error creating in-memory database: %v", err)
	}

	// initially empty
	results, err := db.GetResults(context.Background(), "user")
	if err != nil {
		t.Fatalf("Error getting quiz results: %v", err)
	}

	if results.Correct != 0 {
		t.Fatalf("Expected 0 correct answer, got %d", results.Correct)
	}

	if results.Total != 0 {
		t.Fatalf("Expected 0 total answer, got %d", results.Total)
	}

	// insert one answer
	err = db.InsertQuizAnswer(context.Background(), "user", quiz.QuizAnswer{1: "wrong answer"})
	if err != nil {
		t.Fatalf("Error inserting quiz answer: %v", err)
	}

	// get results back
	results, err = db.GetResults(context.Background(), "user")
	if err != nil {
		t.Fatalf("Error getting quiz results: %v", err)
	}

	if results.Correct != 0 {
		t.Fatalf("Expected 0 correct answer, got %d", results.Correct)
	}

	if results.Total != 1 {
		t.Fatalf("Expected 1 total answer, got %d", results.Total)
	}
}

func TestInsertUser(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Error creating in-memory database: %v", err)
	}

	err = db.InsertUser(context.Background(), "newUser")
	if err != nil {
		t.Fatalf("Error inserting user: %v", err)
	}

	results, err := db.GetResults(context.Background(), "newUser")
	if err != nil {
		t.Fatalf("Error getting quiz results: %v", err)
	}

	if results.Correct != 0 {
		t.Fatalf("Expected 0 correct answer, got %d", results.Correct)
	}

	if results.Total != 0 {
		t.Fatalf("Expected 0 total answer, got %d", results.Total)
	}

	// insert user already exists is err
	err = db.InsertUser(context.Background(), "user")
	if err != ErrUserAlreadyExists {
		t.Fatalf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestGetStatistics(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Error creating in-memory database: %v", err)
	}

	statistics, err := db.GetStatistics(context.Background(), "user")
	if err != ErrNotEnoughUsersForStatistics {
		t.Fatalf("Expected error, not enough users, got %v", err)
	}

	// insert multiple users
	db.InsertUser(context.Background(), "a")
	db.InsertUser(context.Background(), "b")
	db.InsertUser(context.Background(), "c")
	db.InsertUser(context.Background(), "d")

	statistics, err = db.GetStatistics(context.Background(), "user")
	if err != nil {
		t.Fatalf("Error getting statistics: %v", err)
	}

	if statistics.Correct != 0 {
		t.Fatalf("Expected 0 correct answer, got %d", statistics.Correct)
	}

	if statistics.Total != 0 {
		t.Fatalf("Expected 0 total answer, got %d", statistics.Total)
	}

	if statistics.AvgCorrect != 0 {
		t.Fatalf("Expected 0 avg correct answer, got %f", statistics.AvgCorrect)
	}

	if statistics.AvgTotal != 0 {
		t.Fatalf("Expected 0 avg total answer, got %f", statistics.AvgTotal)
	}
}
