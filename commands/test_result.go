package commands

type TestResult struct {
	Action    string `json:"Action"`
	Package   string `json:"Package"`
	Test      string `json:"Test"`
	Passed    bool   `json:"Passed"`
	Output    string `json:"Output"`
	Error     string `json:"Error"`
	Time      string `json:"Time"`
	StartTime string `json:"StartTime"`
	EndTime   string `json:"EndTime"`
}
