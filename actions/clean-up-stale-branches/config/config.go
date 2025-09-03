package config

type Config struct {
	Repository    string
	Owner         string
	DefaultBranch string
	Fetch         bool
	Delete        bool
	CsvFile       string
}
