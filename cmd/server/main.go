//	@title			SHBucket API
//	@version		2.0.0
//	@description	SHBucket is a distributed object storage system similar to AWS S3
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	SHBucket Support
//	@contact.email	support@shbucket.local

//	@license.name	MIT
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-API-Key
//	@description				API Key for authentication

package main

import (
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"

	"shbucket/src/Application/APIKey"
	"shbucket/src/Application/Bucket"
	"shbucket/src/Application/File"
	"shbucket/src/Application/Node"
	"shbucket/src/Application/Setup"
	"shbucket/src/Application/User"
	"shbucket/src/Controllers"
	"shbucket/src/Infrastructure/Auth"
	"shbucket/src/Infrastructure/Mediator"
	"shbucket/src/Infrastructure/Persistence"
	_ "shbucket/docs"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading .env file:", err)
	}
	
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres@localhost:5432/shbucket?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-here"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	log.Println("Connecting to database...")
	dbContext, err := persistence.NewAppDbContext(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbContext.Close()

	log.Println("Database connected successfully")

	
	jwtHandler := auth.NewJWTHandler(jwtSecret, "SHBucket", 24)
	authService := auth.NewAuthorizationService(jwtHandler)
	validator := validator.New()

	// Initialize mediator
	med := mediator.NewMediator()

	// Initialize handlers
	loginHandler := user.NewLoginRequestHandler(dbContext, jwtHandler)
	logoutHandler := user.NewLogoutRequestHandler(dbContext, jwtHandler)
	registerHandler := user.NewRegisterRequestHandler(dbContext)
	changePasswordHandler := user.NewChangePasswordRequestHandler(dbContext)
	getUserHandler := user.NewGetUserRequestHandler(dbContext)
	listUsersHandler := user.NewListUsersRequestHandler(dbContext)

	createBucketHandler := bucket.NewCreateBucketRequestHandler(dbContext)
	deleteBucketHandler := bucket.NewDeleteBucketRequestHandler(dbContext)
	getBucketHandler := bucket.NewGetBucketRequestHandler(dbContext)
	listBucketsHandler := bucket.NewListBucketsRequestHandler(dbContext)
	updateBucketHandler := bucket.NewUpdateBucketRequestHandler(dbContext)

	uploadFileHandler := file.NewUploadFileRequestHandler(dbContext)
	distributedUploadHandler := file.NewDistributedUploadRequestHandler(dbContext)
	deleteFileHandler := file.NewDeleteFileRequestHandler(dbContext)
	getFileHandler := file.NewGetFileRequestHandler(dbContext)
	listFilesHandler := file.NewListFilesRequestHandler(dbContext)
	generateSignedURLHandler := file.NewGenerateSignedURLRequestHandler(dbContext)
	
	createAPIKeyHandler := apikey.NewCreateAPIKeyRequestHandler(dbContext)
	listAPIKeysHandler := apikey.NewListAPIKeysRequestHandler(dbContext)
	deleteAPIKeyHandler := apikey.NewDeleteAPIKeyRequestHandler(dbContext)

	registerNodeHandler := node.NewRegisterNodeRequestHandler(dbContext)
	listNodesHandler := node.NewListNodesRequestHandler(dbContext)

	checkSetupHandler := setup.NewCheckSetupRequestHandler(dbContext)
	masterSetupHandler := setup.NewMasterSetupRequestHandler(dbContext)
	nodeSetupHandler := setup.NewNodeSetupRequestHandler(dbContext)

	// Register handlers with mediator
	med.RegisterHandler(&user.LoginCommand{}, loginHandler)
	med.RegisterHandler(&user.LogoutCommand{}, logoutHandler)
	med.RegisterHandler(&user.RegisterCommand{}, registerHandler)
	med.RegisterHandler(&user.ChangePasswordCommand{}, changePasswordHandler)
	med.RegisterHandler(&user.GetUserCommand{}, getUserHandler)
	med.RegisterHandler(&user.ListUsersCommand{}, listUsersHandler)

	med.RegisterHandler(&bucket.CreateBucketCommand{}, createBucketHandler)
	med.RegisterHandler(&bucket.DeleteBucketCommand{}, deleteBucketHandler)
	med.RegisterHandler(&bucket.GetBucketCommand{}, getBucketHandler)
	med.RegisterHandler(&bucket.ListBucketsCommand{}, listBucketsHandler)
	med.RegisterHandler(&bucket.UpdateBucketCommand{}, updateBucketHandler)

	med.RegisterHandler(&file.UploadFileCommand{}, uploadFileHandler)
	med.RegisterHandler(&file.DistributedUploadCommand{}, distributedUploadHandler)
	med.RegisterHandler(&file.DeleteFileCommand{}, deleteFileHandler)
	med.RegisterHandler(&file.GetFileCommand{}, getFileHandler)
	med.RegisterHandler(&file.ListFilesCommand{}, listFilesHandler)
	med.RegisterHandler(&file.GenerateSignedURLCommand{}, generateSignedURLHandler)
	
	med.RegisterHandler(&apikey.CreateAPIKeyCommand{}, createAPIKeyHandler)
	med.RegisterHandler(&apikey.ListAPIKeysCommand{}, listAPIKeysHandler)
	med.RegisterHandler(&apikey.DeleteAPIKeyCommand{}, deleteAPIKeyHandler)

	med.RegisterHandler(&node.RegisterNodeCommand{}, registerNodeHandler)
	med.RegisterHandler(&node.ListNodesCommand{}, listNodesHandler)

	med.RegisterHandler(&setup.CheckSetupCommand{}, checkSetupHandler)
	med.RegisterHandler(&setup.MasterSetupCommand{}, masterSetupHandler)
	med.RegisterHandler(&setup.NodeSetupCommand{}, nodeSetupHandler)

	// Initialize controllers
	setupController := controllers.NewSetupController(med, validator)
	userController := controllers.NewUserController(med, validator, authService)
	bucketController := controllers.NewBucketController(med, validator, authService)
	fileController := controllers.NewFileController(med, validator, authService, dbContext)
	nodeController := controllers.NewNodeController(med, validator, authService, dbContext)
	apiKeyController := controllers.NewAPIKeyController(med, validator, authService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "SHBucket v2.0.0",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://127.0.0.1:3000",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-API-Key",
	}))


	// Serve static files from web/dist
	app.Static("/", "./web/dist")
	
	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API routes
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// Setup routes (no auth required)
	setup := api.Group("/setup")
	setup.Get("/status", setupController.CheckSetup)
	setup.Post("/master", setupController.SetupMaster)
	setup.Post("/node", setupController.SetupNode)
	setup.Get("/info", setupController.GetSystemInfo)

	// Node self-registration routes (no auth required)
	nodeSetup := api.Group("/node")
	nodeSetup.Post("/register", nodeController.SelfRegister)
	nodeSetup.Get("/auth-key", nodeController.GetAuthKey)
	nodeSetup.Post("/ping", nodeController.Ping)

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", userController.Login)
	auth.Post("/register", userController.Register)
	auth.Post("/logout", authService.RequireRoleOrAPIKey("viewer", dbContext), userController.Logout)
	auth.Post("/change-password", authService.RequireRoleOrAPIKey("viewer", dbContext), userController.ChangePassword)

	// User routes
	users := api.Group("/users", authService.RequireRoleOrAPIKey("admin", dbContext))
	users.Get("/", userController.ListUsers)
	users.Get("/:id", userController.GetUser)

	// Bucket routes
	buckets := api.Group("/buckets", authService.RequireRoleOrAPIKey("viewer", dbContext))
	buckets.Get("/", bucketController.ListBuckets)
	buckets.Post("/", authService.RequireRoleOrAPIKey("editor", dbContext), bucketController.CreateBucket)
	buckets.Put("/:id", authService.RequireRoleOrAPIKey("editor", dbContext), bucketController.UpdateBucket)
	buckets.Get("/:id", bucketController.GetBucket)
	buckets.Delete("/:id", authService.RequireRoleOrAPIKey("manager", dbContext), bucketController.DeleteBucket)

	// File serving route (no auth middleware - handles auth internally)  
	api.Get("/file/:bucketId/:fileId", fileController.ServeFile)
	
	// Internal routes for distributed storage (auth handled internally with node auth key)
	api.Post("/internal/upload", fileController.InternalUpload)
	api.Delete("/internal/delete", fileController.InternalDelete)
	api.Get("/internal/file", fileController.InternalFile)

	// File management routes (require auth)
	files := api.Group("/buckets/:bucketId/files")
	files.Get("/", authService.RequireRoleOrAPIKey("viewer", dbContext), fileController.ListFiles)
	files.Post("/", authService.RequireRoleOrAPIKey("editor", dbContext), fileController.UploadFile)
	files.Get("/:fileId/info", authService.RequireRoleOrAPIKey("viewer", dbContext), fileController.GetFile)  // Metadata only
	files.Delete("/:fileId", authService.RequireRoleOrAPIKey("editor", dbContext), fileController.DeleteFile)
	files.Post("/:fileId/signed-url", authService.RequireRoleOrAPIKey("viewer", dbContext), fileController.GenerateSignedURL)
	
	// API Key routes
	apiKeys := api.Group("/api-keys", authService.RequireRoleOrAPIKey("viewer", dbContext))
	apiKeys.Post("/", apiKeyController.CreateAPIKey)
	apiKeys.Get("/", apiKeyController.ListAPIKeys)
	apiKeys.Delete("/:id", apiKeyController.DeleteAPIKey)

	// Node management routes
	nodes := api.Group("/nodes", authService.RequireRoleOrAPIKey("manager", dbContext))
	nodes.Get("/", nodeController.ListNodes)
	nodes.Post("/", nodeController.RegisterNode)
	nodes.Post("/install", nodeController.InstallNode)
	nodes.Get("/health", nodeController.CheckAllNodesHealth)
	nodes.Get("/:id/health", nodeController.HealthCheck)
	nodes.Delete("/:id", nodeController.DeleteNode)

	// Storage node routes
	storageNodes := api.Group("/storage-nodes", authService.RequireRoleOrAPIKey("manager", dbContext))
	storageNodes.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"storage_nodes": []interface{}{},
			"message":       "Storage nodes not yet implemented",
		})
	})

	// Catch-all route for React Router (SPA)
	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("./web/dist/index.html")
	})

	// Debug: Print all registered routes
	for _, route := range app.GetRoutes() {
		log.Printf("Registered route: %s %s", route.Method, route.Path)
	}

	// Start server
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	log.Printf("SHBucket v2.0 starting on %s:%s", host, port)
	log.Printf("Database: %s", maskDatabaseURL(databaseURL))
	log.Printf("Swagger documentation: http://%s:%s/swagger/", host, port)
	log.Printf("Health check: http://%s:%s/api/v1/health", host, port)

	log.Fatal(app.Listen(host + ":" + port))
}


func maskDatabaseURL(url string) string {
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-7:]
	}
	return "***"
}