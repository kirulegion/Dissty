package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/resend/resend-go/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kirulegion/Dissty/services/auth-service/internal/cache"
	"github.com/kirulegion/Dissty/services/auth-service/internal/handler"
	"github.com/kirulegion/Dissty/services/auth-service/internal/repository"
	"github.com/kirulegion/Dissty/services/auth-service/internal/service"
	"github.com/kirulegion/Dissty/services/auth-service/internal/token"
	userpb "github.com/kirulegion/Dissty/services/user-service/proto"
	pb "github.com/kirulegion/Dissty/services/auth-service/proto"
)

func main() {
	// ── 1. Load config ───────────────────────────────────────────────────────
	grpcPort     := getEnv("GRPC_PORT", ":50052")
	dbURL        := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dissty_auth?sslmode=disable")
	redisAddr    := getEnv("REDIS_ADDR", "localhost:6379")
	resendAPIKey := getEnv("RESEND_API_KEY", "")
	userSvcAddr  := getEnv("USER_SERVICE_ADDR", "localhost:50051")

	// ── 2. Connect to PostgreSQL ─────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to database ✅")

	// ── 3. Connect to Redis ──────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	log.Println("connected to redis ✅")

	// ── 4. Connect to user-service via gRPC ─────────────────────────────────
	userConn, err := grpc.NewClient(userSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to user-service: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserServiceClient(userConn)
	log.Println("connected to user-service ✅")

	// ── 5. Wire up dependencies ──────────────────────────────────────────────
	accountRepo  := repository.NewAccountRepository(db)
	providerRepo := repository.NewIdentityProviderRepository(db)
	otpCache     := cache.NewOTPCache(redisClient)
	resendClient := resend.NewClient(resendAPIKey)
	tokenSvc     := token.NewTokenService()

	authSvc := service.NewAuthService(
		accountRepo,
		providerRepo,
		userClient,
		resendClient,
		otpCache,
		tokenSvc,
	)

	authHandler := handler.NewAuthHandler(authSvc)

	// ── 6. Start gRPC server ─────────────────────────────────────────────────
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	go func() {
		log.Printf("auth-service gRPC server starting on %s 🚀", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// ── 7. Graceful shutdown ─────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down auth-service gracefully...")
	grpcServer.GracefulStop()
	log.Println("auth-service stopped. 👋")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

