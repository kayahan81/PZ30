package main

import (
	"log"
	"net/http"
	"os"

	"tech-ip-sem2/services/graphql/graph"
	"tech-ip-sem2/services/graphql/graph/generated"
	"tech-ip-sem2/services/graphql/internal/repository"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	port := os.Getenv("GRAPHQL_PORT")
	if port == "" {
		port = "8090"
	}

	// Подключение к БД
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "tasks_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "tasks_pass"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "tasks_db"
	}

	// Создаём репозиторий
	repo, err := repository.NewTaskRepository(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer repo.Close()

	// Создаём резолвер
	resolver := &graph.Resolver{Repo: repo}

	// Настраиваем GraphQL сервер
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Роутинг
	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("GraphQL server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
