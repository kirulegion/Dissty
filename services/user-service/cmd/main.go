package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirulegion/Dissty/services/user-service/internal/handler"
	"github.com/kirulegion/Dissty/services/user-service/internal/repository"
	"github.com/kirulegion/Dissty/services/user-service/internal/service"
	pb "github.com/kirulegion/Dissty/services/user-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	//-- 1. Load config from the enviroment variables

	// We read the configuration from the env var.
	grpcPort := getEnv("GRPC_PORT", ":50051")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dissty_users?sslmode=disable")

	//-- 2. Connect to PostgreSQL --------------------------------------------------

	// We create one database connection here and pass it down to the repository.
	// The repository doesn't create it's own connection we inject it here.
	// we create it here and hand it over.

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		// There's no point in starting if we could't connect to the db.
		// INFO: log.Fatalf prints the error and calls os.Exit(1) - hard stop.

		log.Fatalf("failed to connect to the database: %v", err)
	}

	log.Println("Connected to database")

	//-- 3. Wire up the layer of the user service.
	//
	//This is the assembly step - each layer get ht elayer below it injected.
	//Repository get the DB connection.
	//Service get the repository.
	//Handler gets the service.
	//
	//The dependency chain flows inwards
	// handler -> service -> repository -> database.

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// -- 4. Create the gRPC server --------------------------------------------
	//
	// grpc.NewServer() creates a bare gRPC server.
	// Later you'll add interceptors here — for logging, auth, recovery.

	grpcServer := grpc.NewServer()

	// Register our handler with the gRPC server.
	// This tells the server "when a UserService request comes in,
	// send it to userHandler."
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	// reflection allows tools like grpcurl to discover your service's methods.
	// Extremely useful during development — disable in production.
	reflection.Register(grpcServer)

	// -- 5. Start listening ---------------------------------------------------

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcPort, err)
	}

	// Start the gRPC server in a goroutine so it doesn't block.
	// We need the main goroutine free to listen for shutdown signals below.
	go func() {
		log.Printf("user-service gRPC server starting on %s 🚀", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// -- 6. Graceful shutdown -------------------------------------------------
	//
	// We listen for SIGINT (Ctrl+C) or SIGTERM (Docker/Kubernetes stop signal).
	// When received, we stop the gRPC server gracefully —
	// meaning it finishes any in-flight requests before shutting down.
	// Without this, a hard stop could cut off requests mid-execution.

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block here until a shutdown signal is received.
	<-quit

	log.Println("shutting down user-service gracefully...")
	grpcServer.GracefulStop()
	log.Println("user-service stopped. 👋")
}

// getEnv reads an environment variable by key.
// If the variable isn't set, it returns the fallback value.
// This lets us have sensible defaults for local development
// while still being configurable in production.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
