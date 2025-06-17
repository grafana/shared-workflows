package main

import (
	"log"
)

func main() {
	config := getConfigFromEnv()

	analyzer := NewDefaultTestFailureAnalyzer(config)

	if err := analyzer.Run(config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
