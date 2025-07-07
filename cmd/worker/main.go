package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"securities-marketplace/domains/shared/events"
	"securities-marketplace/domains/shared/storage"
)

func main() {
	// Initialize database connection
	db, err := storage.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	redis, err := storage.NewRedisConnection()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Initialize event bus
	eventBus := events.NewEventBus(redis)

	// Initialize event store
	eventStore := events.NewEventStore(db)

	// Start background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start projection workers
	go startProjectionWorkers(ctx, eventStore, eventBus)

	// Start settlement worker
	go startSettlementWorker(ctx, eventStore, eventBus)

	// Start compliance monitoring worker
	go startComplianceWorker(ctx, eventStore, eventBus)

	log.Println("Worker started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	cancel()

	// Give workers time to finish
	time.Sleep(5 * time.Second)

	log.Println("Worker exited")
}

func startProjectionWorkers(ctx context.Context, eventStore events.EventStore, eventBus events.EventBus) {
	log.Println("Starting projection workers...")
	// TODO: Implement projection workers
}

func startSettlementWorker(ctx context.Context, eventStore events.EventStore, eventBus events.EventBus) {
	log.Println("Starting settlement worker...")
	// TODO: Implement settlement worker
}

func startComplianceWorker(ctx context.Context, eventStore events.EventStore, eventBus events.EventBus) {
	log.Println("Starting compliance worker...")
	// TODO: Implement compliance worker
}