package quiz

type Question struct {
	ID      uint64   `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Answer  string   `json:"-"`
}

// nontyped to have something different
type QuizAnswer = map[uint64]string

type User struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
	Correct uint64 `json:"correct"`
	Total   uint64 `json:"total"`
}

// View over User results
type QuizResults struct {
	Correct uint64 `json:"correct"`
	Total   uint64 `json:"total"`
}

type StatisticsResults struct {
	Correct    uint64  `json:"correct"`
	Total      uint64  `json:"total"`
	AvgCorrect float64 `json:"avg_correct"`
	AvgTotal   float64 `json:"avg_total"`
}
