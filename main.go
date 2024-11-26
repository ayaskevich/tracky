package main

import (
	"bufio"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
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

	//Создаём клиент REST API
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
		commits, err := getCommitsWithCurrentAuthor(project.RepoPath, time.Now().AddDate(0, 0, -8), time.Now())
		if err != nil {
			log.Printf("Error processing git project %s: %v", project.Name, err)
			continue
		}

		// Выводим результаты
		if len(commits) == 0 {
			fmt.Printf("No commits found for project %s by the current author\n", project.Name)
		} else {
			fmt.Printf("Found %d commit(s) for project %s by the current author:\n", len(commits), project.Name)
			for _, commit := range commits {
				fmt.Printf("- %s: %s", commit.Author.When, commit.Message)
			}
		}

		// Логируем результаты
		logWorkForProject(project.Name, project.IssueID, commits, config.WorkHours, time.Now().AddDate(0, 0, -10), time.Now(), config, client, project)
	}
}

// isWeekend проверяет, является ли день выходным
func isWeekend(date time.Time) bool {
	weekday := date.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func logWorkForProject(projectName, issueID string, commits []*object.Commit, workHours WorkHours, startPeriod, endPeriod time.Time, config *Config, client *resty.Client, project GitProject) {
	fmt.Printf("\nLogs for project '%s':\n", projectName)
	fmt.Printf("Issue ID: %s\n", issueID)

	//// Парсим рабочие часы
	//startWork, _ := time.Parse("15:04", workHours.Start)
	//endWork, _ := time.Parse("15:04", workHours.End)

	// Текущий анализируемый период
	currentTime := startPeriod

	for i, commit := range commits {
		commitTime := commit.Committer.When

		// Формируем лог для периода с текущего момента до времени коммита
		if currentTime.Before(commitTime) {
			logDuration(currentTime, commitTime, workHours, commit.Message, config, client, project)
			currentTime = commitTime // Обновляем текущий момент
		}

		// Если это последний коммит, закрываем период до конца анализа
		if i == len(commits)-1 && commitTime.Before(endPeriod) {
			logDuration(commitTime, endPeriod, workHours, commit.Message, config, client, project)
		}
	}
}

// logDuration формирует лог работы для периода между start и end
func logDuration(start, end time.Time, workHours WorkHours, description string, config *Config, client *resty.Client, project GitProject) {
	// Идём по дням, считая рабочие часы
	current := start
	for !current.After(end) {
		// Пропускаем выходные
		if isWeekend(current) {
			current = current.AddDate(0, 0, 1)
			continue
		}
		// Начало и конец рабочего дня
		dayStart := time.Date(current.Year(), current.Month(), current.Day(), mustParseHour(workHours.Start).Hour(), mustParseHour(workHours.Start).Minute(), 0, 0, current.Location())
		dayEnd := time.Date(current.Year(), current.Month(), current.Day(), mustParseHour(workHours.End).Hour(), mustParseHour(workHours.End).Minute(), 0, 0, current.Location())

		// Определяем актуальный старт и конец для данного дня
		actualStart := maxTime(start, dayStart)
		actualEnd := minTime(end, dayEnd)

		if actualStart.Before(actualEnd) {
			// Вычисляем продолжительность работы в данном промежутке
			duration := actualEnd.Sub(actualStart).Minutes()
			fmt.Printf("Date: %s, Duration: %.0f minutes, Description: %s\n", actualStart.Format("2006-01-02"), duration, description)

			fmt.Print("Do you want to post this work item? (Y/n): ")
			// Читаем ввод пользователя
			confirmation := askForConfirmation()
			if !confirmation {
				fmt.Println("Skipping this work item...")
				continue
			}

			// Отправляем данные в YouTrack
			err := addWorkItem(client, config.YouTrackURL, config.APIToken, project.IssueID, int(duration), description, actualStart)
			if err != nil {
				log.Printf("Failed to add work item for %s: %v", project.Name, err)
			} else {
				fmt.Printf("Successfully added work item for %s\n", project.Name)
			}

		}

		// Переход на следующий день
		current = current.AddDate(0, 0, 1)
	}
}

// mustParseHour парсит время в формате HH:MM
func mustParseHour(timeStr string) time.Time {
	parsed, err := time.Parse("15:04", timeStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid time format: %s", timeStr))
	}
	return parsed
}

// addWorkItem добавляет запись времени в YouTrack
func addWorkItem(client *resty.Client, youtrackURL, token, issueID string, duration int, description string, actualStart time.Time) error {
	url := fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems", youtrackURL, issueID)

	workItem := map[string]interface{}{
		"date":     actualStart.UnixMilli(),
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
