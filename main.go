package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

func main() {
	// Загружаем конфигурацию из файла
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Создаём клиент REST API
	client := resty.New()

	// Работаем с календарными проектами
	for _, project := range config.CalendarProjects {
		fmt.Printf("Processing calendar project: %s (Issue ID: %s)\n", project.Name, project.IssueID)
		// Логика для обработки календаря
	}

	// Пример: Добавляем запись времени для каждой задачи из конфигурации
	for _, project := range config.GitProjects {
		fmt.Printf("Processing git project: %s (Repo Path: %s, Issue ID: %s)\n", project.Name, project.RepoPath, project.IssueID)
		// Логика для обработки истории Git
		commits, err := getCommitsWithCurrentAuthor(project.RepoPath, time.Now().AddDate(0, 0, -7), time.Now())
		if err != nil {
			log.Printf("Error processing git project %s: %v", project.Name, err)
			continue
		}
		fmt.Printf("Found %d commits for project %s\n", len(commits), project.Name)

		// Выводим результаты
		if len(commits) == 0 {
			fmt.Printf("No commits found for project %s by the current author\n", project.Name)
		} else {
			fmt.Printf("Found %d commit(s) for project %s by the current author:\n", len(commits), project.Name)
			for _, commit := range commits {
				fmt.Printf("- Commit: %s\n  Author: %s\n  Date: %s\n  Message: %s\n\n",
					commit.Hash, commit.Author.Name, commit.Author.When, commit.Message)
			}
		}
		continue
		// Подготавливаем данные для отправки
		duration := 60 // Длительность работы в минутах
		description := fmt.Sprintf("Work on %s", project.Name)

		// Выводим данные для подтверждения пользователю
		fmt.Printf("\nReady to log time for project '%s':\n", project.Name)
		fmt.Printf("Issue ID: %s\n", project.IssueID)
		fmt.Printf("Duration: %d minutes\n", duration)
		fmt.Printf("Description: %s\n", description)
		fmt.Print("Do you want to post this work item? (Y/n): ")

		// Читаем ввод пользователя
		confirmation := askForConfirmation()
		if !confirmation {
			fmt.Println("Skipping this work item...")
			continue
		}

		err = nil
		// Отправляем данные в YouTrack
		err = addWorkItem(client, config.YouTrackURL, config.APIToken, project.IssueID, duration, description)
		if err != nil {
			log.Printf("Failed to add work item for %s: %v", project.Name, err)
		} else {
			fmt.Printf("Successfully added work item for %s\n", project.Name)
		}
	}
}

// addWorkItem добавляет запись времени в YouTrack
func addWorkItem(client *resty.Client, youtrackURL, token, issueID string, duration int, description string) error {
	url := fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems", youtrackURL, issueID)

	workItem := map[string]interface{}{
		"duration": map[string]int{"minutes": duration},
		"text":     description,
	}

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Content-Type", "application/json").
		SetBody(workItem).
		Post(url)

	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("error from YouTrack: %s", resp.String())
	}

	return nil
}

// askForConfirmation запрашивает подтверждение от пользователя
func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading input: %v", err)
		return true // По умолчанию выбираем "yes"
	}

	// Приводим ввод к нижнему регистру и убираем лишние символы
	input = strings.TrimSpace(strings.ToLower(input))

	// Если пользователь ничего не ввёл, возвращаем "yes"
	if input == "" {
		return true
	}

	return input == "y" || input == "yes"
}
