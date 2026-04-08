package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"

    "github.com/joho/godotenv"
    "github.com/kirulegion/Dissty/services/permission-service/internal/handler"
    "github.com/kirulegion/Dissty/services/permission-service/internal/service"
    pb "github.com/kirulegion/Dissty/services/permission-service/proto/permissionpb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

func main() {
    godotenv.Load()

    port := os.Getenv("GRPC_PORT")
    if port == "" {
        port = "50054"
    }

    svc := service.NewPermissionService()
    h := handler.NewPermissionHandler(svc)

    grpcServer := grpc.NewServer()
    pb.RegisterPermissionServiceServer(grpcServer, h)
    reflection.Register(grpcServer)

    lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    go func() {
        log.Printf("permission-service running on :%s", port)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("failed to serve: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("shutting down permission-service")
    grpcServer.GracefulStop()
}
