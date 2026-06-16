package graph

import (
	"tech-ip-sem2/services/graphql/internal/repository"
)

// Resolver — корневой структура для всех резолверов
type Resolver struct {
	Repo *repository.TaskRepository
}
