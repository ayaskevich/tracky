package main

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// getCommitsWithCurrentAuthor получает список коммитов текущего автора из репозитория
func getCommitsWithCurrentAuthor(repoPath string, since, until time.Time) ([]*object.Commit, error) {
	// Открываем локальный репозиторий
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Получаем объект HEAD
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Создаём итератор для коммитов
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	// Получаем автора первого подходящего коммита
	var currentAuthor string
	err = commitIter.ForEach(func(c *object.Commit) error {
		if c.Committer.When.Before(since) {
			return nil
		}
		currentAuthor = c.Author.Name
		return fmt.Errorf("found") // Выходим из цикла, как только найдём автора
	})
	if err != nil && err.Error() != "found" {
		return nil, fmt.Errorf("failed to determine current author: %w", err)
	}

	// Создаём новый итератор для фильтрации по автору
	commitIter, err = repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	// Список коммитов текущего автора
	var filteredCommits []*object.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		// Пропускаем коммиты вне временного диапазона
		if c.Committer.When.Before(since) || c.Committer.When.After(until) {
			return nil
		}

		// Пропускаем коммиты других авторов
		if c.Author.Name != currentAuthor {
			return nil
		}

		// Добавляем коммит текущего автора
		filteredCommits = append(filteredCommits, c)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	return filteredCommits, nil
}
