package graphql

import (
	"context"
	"errors"
	"seta/internal/pkg/auth"
	"seta/internal/pkg/models"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Resolver contains the database connection for GraphQL resolvers.
type Resolver struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewResolver creates a new Resolver.
func NewResolver(db *gorm.DB, log *logrus.Logger) *Resolver {
	return &Resolver{db: db, log: log}
}

// CreateUser creates a new user.
func (r *Resolver) CreateUser(ctx context.Context, args struct {
	Username string
	Email    string
	Password string
	Role     string
}) (*UserResolver, error) {
	if len(args.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters long")
	}

	var existingUser models.User
	if err := r.db.Where("email = ?", args.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("a user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:           uuid.New(),
		Username:     args.Username,
		Email:        args.Email,
		PasswordHash: string(hashedPassword),
		Role:         args.Role,
	}

	if err := r.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &UserResolver{user: user}, nil
}

// Login authenticates a user and returns a token.
// log error when have error
func (r *Resolver) Login(ctx context.Context, args struct {
	Email    string
	Password string
}) (*AuthPayloadResolver, error) {
	var user models.User
	if err := r.db.Where("email = ?", args.Email).First(&user).Error; err != nil {
		r.log.WithFields(logrus.Fields{
			"email": args.Email,
			"error": err.Error(),
		}).Error("Failed to find user by email")
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(args.Password)); err != nil {
		r.log.WithFields(logrus.Fields{
			"email": args.Email,
		}).Error("Invalid password")
		return nil, errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(&user)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate token")
		return nil, err
	}

	return &AuthPayloadResolver{
		token: token,
		user:  user,
	}, nil
}

// Logout logs out a user.
func (r *Resolver) Logout(ctx context.Context) (bool, error) {
	// In a stateless JWT setup, logout is typically handled on the client-side.
	// The server can't invalidate the token, but you could implement a token blacklist here if needed.
	return true, nil
}

// FetchUsers retrieves all users.
func (r *Resolver) FetchUsers(ctx context.Context) ([]*UserResolver, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}

	resolvers := make([]*UserResolver, len(users))
	for i, user := range users {
		resolvers[i] = &UserResolver{user: user}
	}

	return resolvers, nil
}

// UserResolver resolves the User type.
type UserResolver struct {
	user models.User
}

func (r *UserResolver) UserID() graphql.ID {
	return graphql.ID(r.user.ID.String())
}

func (r *UserResolver) Username() string {
	return r.user.Username
}

func (r *UserResolver) Email() string {
	return r.user.Email
}

func (r *UserResolver) Role() string {
	return r.user.Role
}

// AuthPayloadResolver resolves the AuthPayload type.
type AuthPayloadResolver struct {
	token string
	user  models.User
}

func (r *AuthPayloadResolver) Token() string {
	return r.token
}

func (r *AuthPayloadResolver) User() *UserResolver {
	return &UserResolver{user: r.user}
}
