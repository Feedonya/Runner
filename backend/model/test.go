package model

import "encoding/json"

type TestDTO struct {
	Stdin  string `json:"stdin"`
	Stdout string `json:"stdout"`
}

type Test struct {
	ID     int    `json:"id"`
	Stdin  string `json:"stdin"`
	Stdout string `json:"stdout"`
}

type TestResult struct {
	TestNumber int  `json:"test_number"` // or similar
	Passed     bool `json:"passed"`      // or similar
	// Other fields...
}

func ParseTestsJSON(data []byte) ([]TestDTO, error) {
	var tests []TestDTO
	err := json.Unmarshal(data, &tests)
	return tests, err
}
