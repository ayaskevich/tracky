package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config описывает конфигурацию проекта
type Config struct {
	YouTrackURL      string            `json:"youtrack_url"`
	APIToken         string            `json:"api_token"`
	WorkHours        WorkHours         `json:"work_hours"`
	CalendarProjects []CalendarProject `json:"calendar_projects"`
	GitProjects      []GitProject      `json:"git_projects"`
}

// WorkHours описывает рабочие часы
type WorkHours struct {
	Start string `json:"start"` // Время начала рабочего дня (формат HH:MM)
	End   string `json:"end"`   // Время окончания рабочего дня (формат HH:MM)
}

// CalendarProject описывает проект для календарных данных
type CalendarProject struct {
	Name    string `json:"name"`
	IssueID string `json:"issue_id"`
}

// GitProject описывает проект для Git
type GitProject struct {
	Name     string `json:"name"`
	RepoPath string `json:"repo_path"`
	IssueID  string `json:"issue_id"`
}

// LoadConfig загружает конфигурацию из JSON файла
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
