package quiz

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"
)

func TestQuestionMarshalJSON(t *testing.T) {
	question := Question{
		ID:      1,
		Text:    "What is the capital of France?",
		Options: []string{"London", "Paris", "Berlin", "Madrid"},
		Answer:  "Paris",
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(question)
	if err != nil {
		t.Fatalf("Error marshalling question: %v", err)
	}

	got := Question{}
	err = json.NewDecoder(body).Decode(&got)
	if err != nil {
		t.Fatalf("Error unmarshalling question: %v", err)
	}

	if got.ID != question.ID {
		t.Fatalf("Question ID does not match, got: %d, want: %d", got.ID, question.ID)
	}
	if got.Text != question.Text {
		t.Fatalf("Question Text does not match, got: %s, want: %s", got.Text, question.Text)
	}
	if !slices.Equal(got.Options, question.Options) {
		t.Fatalf("Question Options do not match, got: %+v, want: %+v", got.Options, question.Options)
	}
	if got.Answer != "" {
		t.Fatalf("Question Answer does not match, got: %s, want: %s", got.Answer, question.Answer)
	}
}

func TestQuizAnswerMarshalJSON(t *testing.T) {
	answer := QuizAnswer{
		1: "Paris",
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(answer)
	if err != nil {
		t.Fatalf("Error marshalling answers: %v", err)
	}

	got := QuizAnswer{}
	err = json.NewDecoder(body).Decode(&got)
	if err != nil {
		t.Fatalf("Error unmarshalling answers: %v", err)
	}

	if len(got) != len(answer) {
		t.Fatalf("QuizAnswer length does not match, got: %d, want: %d", len(got), len(answer))
	}

	for k, v := range answer {
		if got[k] != v {
			t.Fatalf("QuizAnswer does not match, got: %+v, want: %+v", got, answer)
		}
	}
}

func TestUserMarshalJSON(t *testing.T) {
	user := User{
		ID:      1,
		Name:    "John Doe",
		Correct: 2,
		Total:   3,
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(user)
	if err != nil {
		t.Fatalf("Error marshalling user: %v", err)
	}

	got := User{}
	err = json.NewDecoder(body).Decode(&got)
	if err != nil {
		t.Fatalf("Error unmarshalling user: %v", err)
	}

	if got != user {
		t.Fatalf("User do not match, got: %+v, want: %+v", got, user)
	}
}
