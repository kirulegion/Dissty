package handler

// This is the gRPC handler — the outermost layer of user-service.
// Its only job is:
// 1. Receive gRPC requests
// 2. Convert proto messages → domain types
// 3. Call the service layer
// 4. Convert domain types → proto messages
// 5. Return the response
// It knows nothing about databases, GORM, or business rules.

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kirulegion/Dissty/services/user-service/internal/domain"
	"github.com/kirulegion/Dissty/services/user-service/internal/service"
	pb "github.com/kirulegion/Dissty/services/user-service/proto"
)

// UserHandler implements the gRPC UserServiceServer interface.
// It wraps the service layer and translates between proto and domain types.
type UserHandler struct {
	// This embeds the unimplemented server — a safety net.
	// If you forget to implement a method, Go won't panic —
	// it falls back to the embedded default which returns "not implemented".
	pb.UnimplementedUserServiceServer

	// svc is the service layer — where all business rules live.
	// The handler calls this for every operation.
	svc service.UserService
}

// NewUserHandler is the constructor.
// Takes a UserService interface — not the concrete type.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// CreateUser handles incoming CreateUser gRPC requests.
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Call the service layer with the data from the proto request.
	// The service handles all the rules — uniqueness checks, ID generation etc.
	user, err := h.svc.CreateUser(ctx, req.DisplayName, req.Username, req.Email)
	if err != nil {
		// Convert domain errors → gRPC status errors.
		// gRPC has its own error codes — like HTTP status codes but for RPC.
		return nil, toGRPCError(err)
	}

	// Convert the domain.User → proto User message and return it.
	return &pb.CreateUserResponse{
		User: toProto(user),
	}, nil
}

// GetUserByID handles incoming GetUserByID gRPC requests.
func (h *UserHandler) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.UserResponse, error) {
	// Proto sends IDs as strings — convert back to uuid.UUID.
	id, err := uuid.Parse(req.Id)
	if err != nil {
		// If the ID string isn't a valid UUID, reject immediately.
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	user, err := h.svc.GetUserByID(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UserResponse{
		User: toProto(user),
	}, nil
}

// GetUserByEmail handles incoming GetUserByEmail gRPC requests.
func (h *UserHandler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.UserResponse, error) {
	user, err := h.svc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UserResponse{
		User: toProto(user),
	}, nil
}

// UpdateUser handles incoming UpdateUser gRPC requests.
func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	// Parse the ID from string → uuid.UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	// Build a domain.User from the proto request fields.
	// Only the fields that are allowed to be updated are included here.
	// Email is intentionally excluded — email changes need a separate
	// verification flow (send confirmation email etc.)
	user := &domain.User{
		ID:          id,
		DisplayName: req.DisplayName,
		UserName:    req.Username,
		AvatarURL:   req.AvatarUrl,
		Bio:         req.Bio,
	}

	updated, err := h.svc.UpdateUser(ctx, user)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UserResponse{
		User: toProto(updated),
	}, nil
}

// DeleteUser handles incoming DeleteUser gRPC requests.
func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	if err := h.svc.DeleteUser(ctx, id); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

// toProto converts a domain.User → proto User message.
// This is the translation from internal language → wire language.
// Called every time we need to send a user back over gRPC.
func toProto(u *domain.User) *pb.User {
	return &pb.User{
		Id:          u.ID.String(),
		DisplayName: u.DisplayName,
		Username:    u.UserName,
		Email:       u.Email,
		AvatarUrl:   u.AvatarURL,
		Bio:         u.Bio,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// toGRPCError converts domain errors → gRPC status errors.
// gRPC clients expect status errors with specific codes —
// not raw Go errors. This translation happens in one place
// so every handler method benefits from it automatically.
func toGRPCError(err error) error {
	switch err {
	case domain.ErrUserNotFound:
		// codes.NotFound → equivalent of HTTP 404
		return status.Errorf(codes.NotFound, "user not found")

	case domain.ErrEmailAlreadyExists:
		// codes.AlreadyExists → equivalent of HTTP 409
		return status.Errorf(codes.AlreadyExists, "email already exists")

	case domain.ErrUsernameAlreadyExists:
		return status.Errorf(codes.AlreadyExists, "username already exists")

	case domain.ErrInvalidInput:
		// codes.InvalidArgument → equivalent of HTTP 400
		return status.Errorf(codes.InvalidArgument, "invalid input")

	default:
		// codes.Internal → equivalent of HTTP 500
		// something unexpected happened — don't expose details to the client
		return status.Errorf(codes.Internal, "internal error")
	}
}
