package migrations

import (
	"fmt"
	"os"
	"strings"

	"github.com/shepherrrd/gontext"
	"shbucket/src/Infrastructure/Data/Entities"
)

type MigrationCommands struct {
	ctx     *gontext.DbContext
	manager *gontext.MigrationManager
}

func NewMigrationCommands() (*MigrationCommands, error) {
	ctx, err := CreateSHBucketDesignTimeContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create design-time context: %w", err)
	}

	manager := gontext.NewMigrationManager(ctx, "./migrations", "migrations")

	return &MigrationCommands{
		ctx:     ctx,
		manager: manager,
	}, nil
}

func CreateSHBucketDesignTimeContext() (*gontext.DbContext, error) {
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

func (m *MigrationCommands) AddMigration(migrationName string) error {
	fmt.Printf("🔄 Adding migration: %s\n", migrationName)
	if err := m.manager.AddMigration(migrationName); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	fmt.Printf("✅ Migration '%s' created successfully!\n", migrationName)
	return nil
}

func (m *MigrationCommands) Update() error {
	fmt.Println("🔄 Updating database...")
	if err := m.manager.UpdateDatabase(); err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}
	fmt.Println("✅ Database updated successfully!")
	return nil
}

func (m *MigrationCommands) Status() error {
	fmt.Println("📊 Migration Status")
	fmt.Println("==================")
	if err := m.manager.ListMigrations(); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}
	return nil
}

func (m *MigrationCommands) Rollback(steps int) error {
	fmt.Printf("🔄 Rolling back %d migration(s)...\n", steps)
	fmt.Println("⚠️  WARNING: This will undo schema changes!")
	if err := m.manager.RollbackDatabase(steps); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	fmt.Printf("✅ Successfully rolled back %d migration(s)!\n", steps)
	return nil
}

func (m *MigrationCommands) Drop() error {
	fmt.Println("🗑️  Dropping database...")
	fmt.Println("⚠️  WARNING: This will delete all data!")
	if err := m.manager.DropDatabase(); err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}
	fmt.Println("✅ Database dropped successfully!")
	return nil
}