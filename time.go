package main

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"time"
)

func analyzeCommitTimes(commits []*object.Commit, startPeriod, endPeriod time.Time, workHours WorkHours) []int {
	// Список сессий (в минутах)
	var sessions []int
	if len(commits) == 0 {
		// Если коммитов нет, добавляем всё рабочее время в пределах периода
		totalWorkTime := calculateWorkMinutes(startPeriod, endPeriod, workHours)
		if totalWorkTime > 0 {
			sessions = append(sessions, totalWorkTime)
		}
		return sessions
	}
	//
	//// Парсим рабочие часы
	//startWork, _ := time.Parse("15:04", workHours.Start)
	//endWork, _ := time.Parse("15:04", workHours.End)

	// Первый коммит
	firstCommit := commits[0].Committer.When
	lastCommit := commits[len(commits)-1].Committer.When

	// 1. Учитываем время до первого коммита
	if firstCommit.After(startPeriod) {
		sessions = append(sessions, calculateWorkMinutes(startPeriod, firstCommit, workHours))
	}

	// 2. Учитываем промежутки между коммитами
	for i := 0; i < len(commits)-1; i++ {
		current := commits[i].Committer.When
		next := commits[i+1].Committer.When
		if next.Sub(current) > time.Hour {
			sessions = append(sessions, calculateWorkMinutes(current, next, workHours))
		}
	}

	// 3. Учитываем время после последнего коммита
	if lastCommit.Before(endPeriod) {
		sessions = append(sessions, calculateWorkMinutes(lastCommit, endPeriod, workHours))
	}

	return sessions
}

// calculateWorkMinutes вычисляет рабочие минуты между двумя датами
func calculateWorkMinutes(start, end time.Time, workHours WorkHours) int {
	totalMinutes := 0

	// Парсим рабочие часы
	startWork, _ := time.Parse("15:04", workHours.Start)
	endWork, _ := time.Parse("15:04", workHours.End)

	// Идём по дням от start до end
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		workStart := time.Date(d.Year(), d.Month(), d.Day(), startWork.Hour(), startWork.Minute(), 0, 0, d.Location())
		workEnd := time.Date(d.Year(), d.Month(), d.Day(), endWork.Hour(), endWork.Minute(), 0, 0, d.Location())

		// Если текущий день не пересекается с интервалом, пропускаем
		if end.Before(workStart) || start.After(workEnd) {
			continue
		}

		// Рассчитываем пересечения
		actualStart := maxTime(start, workStart)
		actualEnd := minTime(end, workEnd)
		if actualStart.Before(actualEnd) {
			totalMinutes += int(actualEnd.Sub(actualStart).Minutes())
		}
	}

	return totalMinutes
}

// maxTime возвращает максимальное из двух времён
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// minTime возвращает минимальное из двух времён
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
