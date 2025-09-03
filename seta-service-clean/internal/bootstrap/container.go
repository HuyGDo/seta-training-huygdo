package bootstrap

import (
	"os"
	"seta/internal/adapters/events"
	"seta/internal/adapters/external"
	http "seta/internal/adapters/http/handlers"
	"seta/internal/adapters/repository"
	"seta/internal/application"
	"seta/internal/config"
	"seta/internal/infrastructure"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// AppContainer holds all the dependencies for the application.
type AppContainer struct {
	Log *zerolog.Logger
	DB  *gorm.DB
	Rdb *redis.Client
}

// NewAppContainer creates and initializes a new AppContainer.
func NewAppContainer() (*AppContainer, error) {
	log := infrastructure.NewLogger()
	config.LoadConfig()

	db, err := infrastructure.ConnectDB(log)
	if err != nil {
		return nil, err
	}

	rdb, err := infrastructure.ConnectRedis(log)
	if err != nil {
		return nil, err
	}

	return &AppContainer{Log: log, DB: db, Rdb: rdb}, nil
}

// BuildAndServe builds all dependencies and starts the HTTP server.
func (c *AppContainer) BuildAndServe() {
	// --- Adapters ---
	// Transaction Manager
	txManager := repository.NewGormTransactionManager(c.DB)

	// Repositories
	teamRepo := repository.NewGormTeamRepository(c.DB)
	folderRepo := repository.NewGormFolderRepository(c.DB)
	noteRepo := repository.NewGormNoteRepository(c.DB)
	shareRepo := repository.NewGormShareRepository(c.DB)
	userRepo := repository.NewGormUserRepository(c.DB)

	// Event Publisher
	kafkaPublisher, err := events.NewKafkaPublisher()
	if err != nil {
		c.Log.Fatal().Err(err).Msg("Failed to create Kafka publisher")
	}

	// External Services
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	authService := external.NewGQLAuthService(userServiceURL)
	userImporter := external.NewGQLUserImporter(userServiceURL)

	// --- Application Use Cases ---
	// Team
	createTeamUC := application.NewCreateTeamUseCase(teamRepo, kafkaPublisher, txManager)
	addMemberUC := application.NewAddMemberUseCase(teamRepo, kafkaPublisher)
	removeMemberUC := application.NewRemoveMemberUseCase(teamRepo, kafkaPublisher)
	addManagerUC := application.NewAddManagerUseCase(teamRepo, kafkaPublisher)
	removeManagerUC := application.NewRemoveManagerUseCase(teamRepo, kafkaPublisher)
	getTeamAssetsUC := application.NewGetTeamAssetsUseCase(teamRepo)

	// Folder
	createFolderUC := application.NewCreateFolderUseCase(folderRepo, kafkaPublisher)
	updateFolderUC := application.NewUpdateFolderUseCase(folderRepo, kafkaPublisher)
	createNoteInFolderUC := application.NewCreateNoteInFolderUseCase(folderRepo, noteRepo, kafkaPublisher)
	shareFolderUC := application.NewShareFolderUseCase(folderRepo, shareRepo, kafkaPublisher)
	unshareFolderUC := application.NewUnshareFolderUseCase(folderRepo, shareRepo, kafkaPublisher)

	// Note
	getNoteUC := application.NewGetNoteUseCase(noteRepo)
	updateNoteUC := application.NewUpdateNoteUseCase(noteRepo, kafkaPublisher)
	deleteNoteUC := application.NewDeleteNoteUseCase(noteRepo, kafkaPublisher)
	shareNoteUC := application.NewShareNoteUseCase(noteRepo, shareRepo, kafkaPublisher)
	unshareNoteUC := application.NewUnshareNoteUseCase(noteRepo, shareRepo, kafkaPublisher)

	// User
	importUsersUC := application.NewImportUsersUseCase(userImporter)
	getUserAssetsUC := application.NewGetUserAssetsUseCase(userRepo)

	// --- HTTP Handlers ---
	teamHandler := http.NewTeamHandler(createTeamUC, addMemberUC, removeMemberUC, addManagerUC, removeManagerUC, getTeamAssetsUC)
	folderHandler := http.NewFolderHandler(createFolderUC, updateFolderUC, createNoteInFolderUC, shareFolderUC, unshareFolderUC)
	noteHandler := http.NewNoteHandler(getNoteUC, updateNoteUC, deleteNoteUC, shareNoteUC, unshareNoteUC)
	userHandler := http.NewUserHandler(importUsersUC, getUserAssetsUC)

	// --- HTTP Server ---
	router := SetupRouter(c.Log, authService, teamRepo, teamHandler, folderHandler, noteHandler, userHandler)
	c.Log.Info().Msg("Starting server on port 8080")
	if err := router.Run(":8080"); err != nil {
		c.Log.Fatal().Err(err).Msg("Could not start server")
	}
}
