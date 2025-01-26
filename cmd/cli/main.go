package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/vrnvu/temp/pkg/quiz"
)

const (
	apiURL             = "https://localhost:8080"
	pathGetQuiz        = "quiz"
	pathGetQuizResults = "quiz/%s"
	pathPutQuizAnswer  = "quiz/%s"
	pathGetStatistics  = "statistics/%s"
)

const usage = `
Quiz CLI
Usage:
	cli --user <token> <command>

Commands:
	quiz      Take a quiz
	results   Show quiz results
	statistics Show statistics
Example:
	cli --user user quiz
	cli --user user results
	cli --user user statistics
`

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }

	var userKey string
	flag.StringVar(&userKey, "user", "", "User authentication token")
	flag.StringVar(&userKey, "u", "", "User authentication token")
	flag.Parse()

	if userKey == "" {
		logger.Error("Error: --user is required")
		flag.Usage()
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 {
		logger.Error("Error: command is required")
		flag.Usage()
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "quiz":
		runQuiz(userKey)
	case "results":
		showResults(userKey)
	case "statistics":
		showStatistics(userKey)
	default:
		logger.Error("Unknown command", "command", command)
		flag.Usage()
		os.Exit(1)
	}
}

func newHTTPSClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func runQuiz(userKey string) {
	client := newHTTPSClient()
	url := fmt.Sprintf("%s/%s", apiURL, pathGetQuiz)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error getting questions: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var questions []quiz.Question
	err = json.NewDecoder(resp.Body).Decode(&questions)
	if err != nil {
		fmt.Printf("Error decoding questions: %v\n", err)
		return
	}

	answers := make(map[uint64]string)
	for _, q := range questions {
		prompt := promptui.Select{
			Label: q.Text,
			Items: q.Options,
		}

		_, value, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		answers[q.ID] = value
	}

	body := bytes.NewBuffer(nil)
	err = json.NewEncoder(body).Encode(answers)
	if err != nil {
		fmt.Printf("Error marshalling answers: %v\n", err)
		return
	}

	url = fmt.Sprintf("%s/%s", apiURL, fmt.Sprintf(pathPutQuizAnswer, userKey))
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Error submitting answers: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("\nAnswers submitted successfully!")
}

func showResults(userKey string) {
	client := newHTTPSClient()
	url := fmt.Sprintf("%s/%s", apiURL, fmt.Sprintf(pathGetQuizResults, userKey))
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			fmt.Println("User not found")
		case http.StatusBadRequest:
			fmt.Println("bad request")
		default:
			fmt.Printf("Error getting results: %v\n", err)
		}
		return
	}
	defer resp.Body.Close()

	var results quiz.QuizResults
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		fmt.Printf("Error decoding results: %v\n", err)
		return
	}

	fmt.Println(results)
}

func showStatistics(userKey string) {
	client := newHTTPSClient()
	url := fmt.Sprintf("%s/%s", apiURL, fmt.Sprintf(pathGetStatistics, userKey))
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			fmt.Println("user not found")
		case http.StatusBadRequest:
			fmt.Println("bad request, not enough users for statistics")
		default:
			fmt.Printf("Error getting statistics: %v\n", err)
		}
		return
	}

	var statistics quiz.StatisticsResults
	err = json.NewDecoder(resp.Body).Decode(&statistics)
	if err != nil {
		fmt.Printf("Error decoding statistics: %v\n", err)
		return
	}

	fmt.Println(statistics)
}
