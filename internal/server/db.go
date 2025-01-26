package server

import (
	"context"
	"errors"
	"math/rand"
	"sync"

	"github.com/vrnvu/temp/pkg/quiz"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrNotEnoughUsersForStatistics = errors.New("not enough users for statistics")

type InMemoryDB struct {
	questions     []quiz.Question
	lockQuestions sync.RWMutex
	users         []quiz.User
	lockUsers     sync.RWMutex
}

func NewInMemoryDB() (*InMemoryDB, error) {
	questions := []quiz.Question{
		{ID: 0, Text: "What is the capital of France?", Options: []string{"London", "Paris", "Berlin", "Madrid"}, Answer: "Paris"},
		{ID: 1, Text: "What is the capital of Germany?", Options: []string{"Berlin", "Paris", "London", "Madrid"}, Answer: "Berlin"},
		{ID: 2, Text: "What is 2 + 2?", Options: []string{"1", "2", "3", "4"}, Answer: "4"},
		{ID: 3, Text: "What is 2 * 2?", Options: []string{"1", "2", "3", "4"}, Answer: "4"},
		{ID: 4, Text: "What is 2 - 2?", Options: []string{"0", "1", "2", "3"}, Answer: "0"},
	}

	users := []quiz.User{
		{ID: 0, Name: "user", Correct: 0, Total: 0},
	}

	return &InMemoryDB{
		questions: questions,
		users:     users,
	}, nil
}

func (db *InMemoryDB) GetQuestions(_ context.Context) ([]quiz.Question, error) {
	db.lockQuestions.RLock()
	defer db.lockQuestions.RUnlock()

	i1 := rand.Intn(len(db.questions))
	i2 := rand.Intn(len(db.questions))
	return []quiz.Question{db.questions[i1], db.questions[i2]}, nil
}

func (db *InMemoryDB) InsertQuizAnswer(_ context.Context, user string, answer quiz.QuizAnswer) error {
	db.lockQuestions.RLock()
	defer db.lockQuestions.RUnlock()

	db.lockUsers.RLock()
	defer db.lockUsers.RUnlock()

	userID, err := db.getUserID(user)
	if err != nil {
		return err
	}

	for questionID, userAnswer := range answer {
		if db.questions[questionID].Answer == userAnswer {
			db.users[userID].Correct++
		}
		db.users[userID].Total++
	}

	return nil
}

func (db *InMemoryDB) GetResults(_ context.Context, user string) (quiz.QuizResults, error) {
	db.lockUsers.RLock()
	defer db.lockUsers.RUnlock()

	userID, err := db.getUserID(user)
	if err != nil {
		return quiz.QuizResults{}, err
	}

	userResults := quiz.QuizResults{
		Correct: db.users[userID].Correct,
		Total:   db.users[userID].Total,
	}

	return userResults, nil
}

func (db *InMemoryDB) getUserID(username string) (uint64, error) {
	for _, u := range db.users {
		if u.Name == username {
			return u.ID, nil
		}
	}

	return 0, ErrUserNotFound
}

func (db *InMemoryDB) InsertUser(_ context.Context, user string) error {
	if _, err := db.getUserID(user); err == nil {
		return ErrUserAlreadyExists
	}

	db.lockUsers.Lock()
	defer db.lockUsers.Unlock()

	userID := uint64(len(db.users))
	db.users = append(db.users, quiz.User{ID: userID, Name: user, Correct: 0, Total: 0})
	return nil
}

func (db *InMemoryDB) GetStatistics(_ context.Context, userName string) (quiz.StatisticsResults, error) {
	if len(db.users) < 2 {
		return quiz.StatisticsResults{}, ErrNotEnoughUsersForStatistics
	}

	userID, err := db.getUserID(userName)
	if err != nil {
		return quiz.StatisticsResults{}, err
	}

	user := db.users[userID]

	statisticsCorrect := uint64(0)
	statisticsTotal := uint64(0)
	for _, user := range db.users {
		// since names must be unique
		if user.Name == userName {
			continue
		}
		statisticsCorrect += user.Correct
		statisticsTotal += user.Total
	}

	avgCorrect := float64(statisticsCorrect) / float64(len(db.users)-1)
	avgTotal := float64(statisticsTotal) / float64(len(db.users)-1)

	return quiz.StatisticsResults{Correct: user.Correct, Total: user.Total, AvgCorrect: avgCorrect, AvgTotal: avgTotal}, nil
}
