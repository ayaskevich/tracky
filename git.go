package main

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"sort"
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

	// Получаем список всех рефов (веток)
	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	// Список уникальных хэшей коммитов
	commitHashes := map[string]bool{}
	var filteredCommits []*object.Commit
	var currentAuthor string

	// Итерация по всем рефам (веткам)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Пропускаем теги или другие рефы, если это не ветки
		if !ref.Name().IsBranch() {
			return nil
		}

		// Получаем лог для ветки
		commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return fmt.Errorf("failed to get commit log for branch %s: %w", ref.Name(), err)
		}

		// Перебираем коммиты
		err = commitIter.ForEach(func(c *object.Commit) error {
			// Пропускаем коммиты вне временного диапазона
			if c.Committer.When.Before(since) || c.Committer.When.After(until) {
				return nil
			}

			// Устанавливаем текущего автора из первого подходящего коммита
			if currentAuthor == "" {
				currentAuthor = c.Author.Name
			}

			// Пропускаем коммиты других авторов
			if c.Author.Name != currentAuthor {
				return nil
			}

			// Добавляем коммит, если он ещё не был обработан
			if !commitHashes[c.Hash.String()] {
				commitHashes[c.Hash.String()] = true
				filteredCommits = append(filteredCommits, c)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("error iterating commits: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process references: %w", err)
	}

	// Сортировка коммитов по дате (от старых к новым)
	sort.Slice(filteredCommits, func(i, j int) bool {
		return filteredCommits[i].Committer.When.Before(filteredCommits[j].Committer.When)
	})

	return filteredCommits, nil
}
