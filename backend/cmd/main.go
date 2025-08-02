package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/ringecosystem/degov-apps/graph"
	"github.com/ringecosystem/degov-apps/internal"
	"github.com/ringecosystem/degov-apps/internal/config"
	"github.com/ringecosystem/degov-apps/internal/directives"
	"github.com/ringecosystem/degov-apps/internal/middleware"
	"github.com/ringecosystem/degov-apps/tasks"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	// Initialize the application
	internal.AppInit()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background tasks
	startBackgroundTasks(ctx)

	// Handle graceful shutdown
	go handleGracefulShutdown(cancel)

	// Start the web server
	startServer()
}

// startBackgroundTasks starts all background tasks
func startBackgroundTasks(ctx context.Context) {
	slog.Info("Starting background tasks...")

	// Create task manager
	taskManager, err := tasks.NewTaskManager()
	if err != nil {
		slog.Error("Failed to create task manager", "error", err)
		return
	}

	// Get task definitions (combines config and constructor)
	taskDefinitions := tasks.GetTaskDefinitions()

	// Register tasks based on definitions
	registeredCount := 0
	for _, def := range taskDefinitions {
		if !def.Config.Enabled {
			slog.Info("Task disabled, skipping", "task", def.Config.Name)
			continue
		}

		task := def.Constructor()
		if err := taskManager.RegisterTask(task, def.Config.Interval); err != nil {
			slog.Error("Failed to register task", "task", def.Config.Name, "error", err)
			continue
		}

		registeredCount++
	}

	slog.Info("Registered tasks",
		"total_count", registeredCount,
		"registered_tasks", taskManager.ListTasks())

	// Start the task manager
	go taskManager.Start(ctx)

	slog.Info("Background tasks started successfully")
}

// handleGracefulShutdown handles graceful shutdown signals
func handleGracefulShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	slog.Info("Received shutdown signal, gracefully shutting down...")
	cancel()
	time.Sleep(2 * time.Second) // Give time for background tasks to finish
	os.Exit(0)
}

// startServer starts the GraphQL server
func startServer() {
	cfg := config.GetConfig()
	port := cfg.GetPort()

	// Configure directives
	graphqlConfig := graph.Config{
		Resolvers: graph.NewResolver(),
		Directives: graph.DirectiveRoot{
			Auth:      directives.AuthDirective,
			Authorize: directives.AuthorizeDirective,
		},
	}

	srv := handler.New(graph.NewExecutableSchema(graphqlConfig))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Initialize authentication middleware
	authMiddleware := middleware.NewAuthMiddleware()

	mux := http.NewServeMux()

	graphiql := playground.Handler("GraphQL playground", "/graphql", playground.WithGraphiqlEnablePluginExplorer(true))
	mux.Handle("/graphiql", graphiql)
	mux.Handle("/graphql", authMiddleware.HTTPMiddleware(srv))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            config.GetAppEnv().IsDevelopment(),
	})
	httpHandler := corsHandler.Handler(mux)

	slog.Info(
		"Server is running",
		slog.String("listen", "http://::"+port+"/"),
	)
	err := http.ListenAndServe(":"+port, httpHandler)
	slog.Error("failed to listen server", "error", err)
}
