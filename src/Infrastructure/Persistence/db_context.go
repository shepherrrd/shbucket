package persistence

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shepherrrd/gontext"
	"shbucket/src/Infrastructure/Data/Entities"
)

type AppDbContext struct {
	*gontext.DbContext
	
	Users            *gontext.LinqDbSet[entities.User]
	Sessions         *gontext.LinqDbSet[entities.Session]
	Buckets          *gontext.LinqDbSet[entities.Bucket]
	Files            *gontext.LinqDbSet[entities.File]
	StorageNodes     *gontext.LinqDbSet[entities.StorageNode]
	APIKeys          *gontext.LinqDbSet[entities.APIKey]
	SignedURLs       *gontext.LinqDbSet[entities.SignedURL]
	SetupConfigs     *gontext.LinqDbSet[entities.SetupConfig]
	NodeFileMetadata *gontext.LinqDbSet[entities.NodeFileMetadata]
}

func NewAppDbContext(databaseURL string) (*AppDbContext, error) {
	logLevel := "info"
	if envLevel := os.Getenv("DB_LOG_LEVEL"); envLevel != "" {
		logLevel = envLevel
	}
	
	ctx, err := gontext.NewDbContext(databaseURL, "postgres", logLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create GoNtext context: %w", err)
	}

	users := gontext.RegisterEntity[entities.User](ctx)
	sessions := gontext.RegisterEntity[entities.Session](ctx)
	buckets := gontext.RegisterEntity[entities.Bucket](ctx)
	files := gontext.RegisterEntity[entities.File](ctx)
	storageNodes := gontext.RegisterEntity[entities.StorageNode](ctx)
	apiKeys := gontext.RegisterEntity[entities.APIKey](ctx)
	signedURLs := gontext.RegisterEntity[entities.SignedURL](ctx)
	setupConfigs := gontext.RegisterEntity[entities.SetupConfig](ctx)
	nodeFileMetadata := gontext.RegisterEntity[entities.NodeFileMetadata](ctx)

	sqlDB, err := ctx.GetDB().DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return &AppDbContext{
		DbContext:        ctx,
		Users:            users,
		Sessions:         sessions,
		Buckets:          buckets,
		Files:            files,
		StorageNodes:     storageNodes,
		APIKeys:          apiKeys,
		SignedURLs:       signedURLs,
		SetupConfigs:     setupConfigs,
		NodeFileMetadata: nodeFileMetadata,
	}, nil
}

func CreateDesignTimeContext() (*gontext.DbContext, error) {
	connectionString := "postgres://postgres@localhost:5432/shbucket?sslmode=disable"
	
	if envURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); envURL != "" {
		connectionString = envURL
	}

	ctx, err := gontext.NewDbContext(connectionString, "postgres")
	if err != nil {
		return nil, err
	}

	gontext.RegisterEntity[entities.User](ctx)
	gontext.RegisterEntity[entities.Session](ctx)
	gontext.RegisterEntity[entities.Bucket](ctx)
	gontext.RegisterEntity[entities.File](ctx)
	gontext.RegisterEntity[entities.StorageNode](ctx)
	gontext.RegisterEntity[entities.APIKey](ctx)
	gontext.RegisterEntity[entities.SignedURL](ctx)
	gontext.RegisterEntity[entities.SetupConfig](ctx)
	gontext.RegisterEntity[entities.NodeFileMetadata](ctx)

	return ctx, nil
}

