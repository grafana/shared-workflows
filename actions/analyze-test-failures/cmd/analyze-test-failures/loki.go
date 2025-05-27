package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LokiClient interface {
	FetchLogs() (*LokiResponse, error)
}

type DefaultLokiClient struct {
	config Config
}

type LokiResponse struct {
	Status string   `json:"status"`
	Data   LokiData `json:"data"`
}

type LokiData struct {
	ResultType string       `json:"resultType"`
	Result     []LokiResult `json:"result"`
}

type LokiResult struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func NewDefaultLokiClient(config Config) *DefaultLokiClient {
	return &DefaultLokiClient{config: config}
}

func (l *DefaultLokiClient) FetchLogs() (*LokiResponse, error) {
	logsJSON, err := fetchLogsFromLoki(l.config)
	if err != nil {
		return nil, err
	}

	var lokiResp LokiResponse
	if err := json.Unmarshal([]byte(logsJSON), &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response: %w", err)
	}

	return &lokiResp, nil
}

func fetchLogsFromLoki(config Config) (string, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/loki/api/v1/query_range", config.LokiURL), nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	endTime := time.Now()
	startTime, err := parseTimeRange(config.TimeRange, endTime)
	if err != nil {
		return "", fmt.Errorf("parsing time range: %w", err)
	}

	query := buildLogQLQuery(config.Repository, config.TimeRange)

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", fmt.Sprintf("%d", startTime.UnixNano()))
	q.Add("end", fmt.Sprintf("%d", endTime.UnixNano()))
	q.Add("limit", "5000")
	req.URL.RawQuery = q.Encode()

	if config.LokiUsername != "" && config.LokiPassword != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(config.LokiUsername + ":" + config.LokiPassword))
		req.Header.Add("Authorization", "Basic "+auth)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bad response from Loki (status %d): %s",
			resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	var tempResp LokiResponse
	if err := json.Unmarshal(body, &tempResp); err == nil {
		totalEntries := 0
		for _, result := range tempResp.Data.Result {
			totalEntries += len(result.Values)
		}
		log.Printf("📥 Retrieved %d log entries from %d streams", totalEntries, len(tempResp.Data.Result))
	}

	return string(body), nil
}

func parseTimeRange(timeRange string, endTime time.Time) (time.Time, error) {
	if strings.HasSuffix(timeRange, "d") {
		daysStr := strings.TrimSuffix(timeRange, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid day format: %s", timeRange)
		}
		duration := time.Duration(days) * 24 * time.Hour
		return endTime.Add(-duration), nil
	}

	duration, err := time.ParseDuration(timeRange)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time range format: %s (supported: 1h, 30m, 7d, etc.)", timeRange)
	}

	return endTime.Add(-duration), nil
}

func buildLogQLQuery(repository, timeRange string) string {
	return fmt.Sprintf(`{service_name="%s", service_namespace="cicd-o11y"} |= "--- FAIL: Test" | json | __error__="" | resources_ci_github_workflow_run_conclusion!="cancelled" | line_format "{{.body}}" | regexp "--- FAIL: (?P<test_name>.*) \\(\\d" | line_format "{{.test_name}}" | regexp `+"`(?P<parent_test_name>Test[a-z0-9A-Z_]+)`", repository)
}
