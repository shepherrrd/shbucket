// Code generated migration. DO NOT EDIT.
package migrations

import (
	"gorm.io/gorm"
)

type Migration20250816103043 struct{}

func (m *Migration20250816103043) ID() string {
	return "20250816103043_initialcreate"
}

func (m *Migration20250816103043) Up(db *gorm.DB) error {
	// Create table StorageNode
	if err := db.Exec("CREATE TABLE \"StorageNode\" (\"UpdatedAt\" TIMESTAMP NOT NULL, \"Id\" UUID NOT NULL DEFAULT gen_random_uuid(), \"IsActive\" BOOLEAN NOT NULL DEFAULT true, \"IsHealthy\" BOOLEAN NOT NULL DEFAULT false, \"MaxStorage\" BIGINT NOT NULL DEFAULT 0, \"UsedStorage\" BIGINT NOT NULL DEFAULT 0, \"CreatedAt\" TIMESTAMP NOT NULL, \"LastPing\" TIMESTAMP, \"Name\" TEXT NOT NULL, \"URL\" TEXT NOT NULL, \"AuthKey\" TEXT NOT NULL, \"Priority\" INTEGER NOT NULL DEFAULT 0, PRIMARY KEY (\"Id\"), CONSTRAINT \"uni_StorageNode_URL\" UNIQUE (\"URL\"))").Error; err != nil {
		return err
	}
	// Create table SetupConfig
	if err := db.Exec("CREATE TABLE \"SetupConfig\" (\"ConfigData\" TEXT, \"CreatedAt\" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, \"UpdatedAt\" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, \"NodeName\" TEXT NOT NULL, \"StoragePath\" TEXT NOT NULL, \"MaxStorage\" BIGINT NOT NULL DEFAULT 0, \"ID\" UUID NOT NULL DEFAULT gen_random_uuid(), \"IsSetup\" BOOLEAN NOT NULL DEFAULT false, \"SetupType\" TEXT NOT NULL, \"MasterURL\" TEXT NOT NULL, PRIMARY KEY (\"ID\"))").Error; err != nil {
		return err
	}
	// Create table User
	if err := db.Exec("CREATE TABLE \"User\" (\"Id\" UUID NOT NULL DEFAULT gen_random_uuid(), \"Username\" TEXT NOT NULL, \"Email\" TEXT NOT NULL, \"PasswordHash\" TEXT NOT NULL, \"CreatedAt\" TIMESTAMP NOT NULL, \"UpdatedAt\" TIMESTAMP NOT NULL, \"LastLoginTime\" TIMESTAMP, \"Buckets\" TEXT, \"Role\" TEXT NOT NULL DEFAULT 'viewer', \"IsActive\" BOOLEAN NOT NULL DEFAULT true, \"PhoneNumber\" TEXT, \"Sessions\" TEXT, PRIMARY KEY (\"Id\"), CONSTRAINT \"uni_User_Username\" UNIQUE (\"Username\"), CONSTRAINT \"uni_User_Email\" UNIQUE (\"Email\"))").Error; err != nil {
		return err
	}
	// Create table Bucket
	if err := db.Exec("CREATE TABLE \"Bucket\" (\"Id\" UUID NOT NULL DEFAULT gen_random_uuid(), \"Name\" TEXT NOT NULL, \"Description\" TEXT NOT NULL, \"Settings\" TEXT NOT NULL, \"CreatedAt\" TIMESTAMP NOT NULL, \"UpdatedAt\" TIMESTAMP NOT NULL, \"OwnerId\" UUID NOT NULL, \"Owner\" TEXT NOT NULL, \"AuthRule\" TEXT NOT NULL, \"Files\" TEXT, PRIMARY KEY (\"Id\"), CONSTRAINT \"uni_Bucket_Name\" UNIQUE (\"Name\"), CONSTRAINT \"fk_Bucket_Id\" FOREIGN KEY (\"Id\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_Bucket_Name\" FOREIGN KEY (\"Name\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_Bucket_Settings\" FOREIGN KEY (\"Settings\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_Bucket_CreatedAt\" FOREIGN KEY (\"CreatedAt\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_Bucket_UpdatedAt\" FOREIGN KEY (\"UpdatedAt\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_Bucket_OwnerId\" FOREIGN KEY (\"OwnerId\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_Bucket_Owner\" FOREIGN KEY (\"Owner\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_Bucket_AuthRule\" FOREIGN KEY (\"AuthRule\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_Bucket_Files\" FOREIGN KEY (\"Files\") REFERENCES \"User\" (\"Id\"))").Error; err != nil {
		return err
	}
	// Create table APIKey
	if err := db.Exec("CREATE TABLE \"APIKey\" (\"Id\" UUID NOT NULL DEFAULT gen_random_uuid(), \"Name\" TEXT NOT NULL, \"KeyPrefix\" TEXT NOT NULL, \"UserId\" UUID NOT NULL, \"Permissions\" TEXT, \"CreatedAt\" TIMESTAMP NOT NULL, \"UpdatedAt\" TIMESTAMP NOT NULL, \"KeyHash\" TEXT NOT NULL, \"IsActive\" BOOLEAN NOT NULL DEFAULT true, \"ExpiresAt\" TIMESTAMP, \"LastUsed\" TIMESTAMP, \"User\" TEXT NOT NULL, PRIMARY KEY (\"Id\"), CONSTRAINT \"uni_APIKey_KeyHash\" UNIQUE (\"KeyHash\"), CONSTRAINT \"fk_APIKey_Id\" FOREIGN KEY (\"Id\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_APIKey_Name\" FOREIGN KEY (\"Name\") REFERENCES \"Session\" (\"Id\"), CONSTRAINT \"fk_APIKey_KeyPrefix\" FOREIGN KEY (\"KeyPrefix\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_APIKey_UserId\" FOREIGN KEY (\"UserId\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_APIKey_Permissions\" FOREIGN KEY (\"Permissions\") REFERENCES \"Session\" (\"Id\"), CONSTRAINT \"fk_APIKey_CreatedAt\" FOREIGN KEY (\"CreatedAt\") REFERENCES \"Session\" (\"Id\"), CONSTRAINT \"fk_APIKey_UpdatedAt\" FOREIGN KEY (\"UpdatedAt\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_APIKey_KeyHash\" FOREIGN KEY (\"KeyHash\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_APIKey_IsActive\" FOREIGN KEY (\"IsActive\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_APIKey_User\" FOREIGN KEY (\"User\") REFERENCES \"APIKey\" (\"Id\"))").Error; err != nil {
		return err
	}
	// Create table NodeFileMetadata
	if err := db.Exec("CREATE TABLE \"NodeFileMetadata\" (\"Id\" UUID NOT NULL DEFAULT gen_random_uuid(), \"BucketId\" UUID NOT NULL, \"BucketName\" TEXT NOT NULL, \"Filename\" TEXT NOT NULL, \"Path\" TEXT NOT NULL, \"Size\" BIGINT NOT NULL, \"CreatedAt\" TIMESTAMP NOT NULL, CONSTRAINT \"fk_NodeFileMetadata_Id\" FOREIGN KEY (\"Id\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_NodeFileMetadata_BucketId\" FOREIGN KEY (\"BucketId\") REFERENCES \"File\" (\"Id\"), CONSTRAINT \"fk_NodeFileMetadata_BucketName\" FOREIGN KEY (\"BucketName\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_NodeFileMetadata_Filename\" FOREIGN KEY (\"Filename\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_NodeFileMetadata_Path\" FOREIGN KEY (\"Path\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_NodeFileMetadata_Size\" FOREIGN KEY (\"Size\") REFERENCES \"File\" (\"Id\"), CONSTRAINT \"fk_NodeFileMetadata_CreatedAt\" FOREIGN KEY (\"CreatedAt\") REFERENCES \"Bucket\" (\"Id\"))").Error; err != nil {
		return err
	}
	// Create table File
	if err := db.Exec("CREATE TABLE \"File\" (\"Id\" UUID NOT NULL DEFAULT gen_random_uuid(), \"BucketId\" UUID NOT NULL, \"Bucket\" TEXT NOT NULL, \"SecuredUrl\" TEXT NOT NULL, \"OriginalName\" TEXT NOT NULL, \"Size\" BIGINT NOT NULL, \"MimeType\" TEXT NOT NULL, \"AccessedAt\" TIMESTAMP, \"UploadedBy\" UUID NOT NULL, \"CreatedAt\" TIMESTAMP NOT NULL, \"Name\" TEXT NOT NULL, \"Path\" TEXT NOT NULL, \"Checksum\" TEXT NOT NULL, \"AuthRule\" TEXT NOT NULL, \"Metadata\" TEXT NOT NULL, \"Version\" INTEGER NOT NULL DEFAULT 1, \"UpdatedAt\" TIMESTAMP NOT NULL, PRIMARY KEY (\"Id\"), CONSTRAINT \"fk_File_Id\" FOREIGN KEY (\"Id\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_BucketId\" FOREIGN KEY (\"BucketId\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_Bucket\" FOREIGN KEY (\"Bucket\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_SecuredUrl\" FOREIGN KEY (\"SecuredUrl\") REFERENCES \"File\" (\"Id\"), CONSTRAINT \"fk_File_OriginalName\" FOREIGN KEY (\"OriginalName\") REFERENCES \"File\" (\"Id\"), CONSTRAINT \"fk_File_Size\" FOREIGN KEY (\"Size\") REFERENCES \"File\" (\"Id\"), CONSTRAINT \"fk_File_MimeType\" FOREIGN KEY (\"MimeType\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_UploadedBy\" FOREIGN KEY (\"UploadedBy\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_CreatedAt\" FOREIGN KEY (\"CreatedAt\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_Name\" FOREIGN KEY (\"Name\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_Path\" FOREIGN KEY (\"Path\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_Checksum\" FOREIGN KEY (\"Checksum\") REFERENCES \"File\" (\"Id\"), CONSTRAINT \"fk_File_AuthRule\" FOREIGN KEY (\"AuthRule\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_Metadata\" FOREIGN KEY (\"Metadata\") REFERENCES \"Bucket\" (\"Id\"), CONSTRAINT \"fk_File_Version\" FOREIGN KEY (\"Version\") REFERENCES \"File\" (\"Id\"), CONSTRAINT \"fk_File_UpdatedAt\" FOREIGN KEY (\"UpdatedAt\") REFERENCES \"Bucket\" (\"Id\"))").Error; err != nil {
		return err
	}
	// Create table SignedURL
	if err := db.Exec("CREATE TABLE \"SignedURL\" (\"BucketName\" TEXT NOT NULL, \"FileName\" TEXT NOT NULL, \"Used\" BOOLEAN NOT NULL DEFAULT false, \"UsedAt\" TIMESTAMP, \"Method\" TEXT NOT NULL, \"ExpiresAt\" TIMESTAMP NOT NULL, \"CreatedAt\" TIMESTAMP NOT NULL, \"SingleUse\" BOOLEAN NOT NULL DEFAULT false, \"ID\" UUID NOT NULL DEFAULT gen_random_uuid(), \"Signature\" TEXT NOT NULL, PRIMARY KEY (\"ID\"), CONSTRAINT \"uni_SignedURL_Signature\" UNIQUE (\"Signature\"))").Error; err != nil {
		return err
	}
	// Create table Session
	if err := db.Exec("CREATE TABLE \"Session\" (\"IsActive\" BOOLEAN NOT NULL DEFAULT true, \"ExpiresAt\" TIMESTAMP NOT NULL, \"CreatedAt\" TIMESTAMP NOT NULL, \"LastUsed\" TIMESTAMP NOT NULL, \"Id\" UUID NOT NULL DEFAULT gen_random_uuid(), \"UserId\" UUID NOT NULL, \"User\" TEXT NOT NULL, \"TokenHash\" TEXT NOT NULL, PRIMARY KEY (\"Id\"), CONSTRAINT \"uni_Session_TokenHash\" UNIQUE (\"TokenHash\"), CONSTRAINT \"fk_Session_IsActive\" FOREIGN KEY (\"IsActive\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_Session_ExpiresAt\" FOREIGN KEY (\"ExpiresAt\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_Session_CreatedAt\" FOREIGN KEY (\"CreatedAt\") REFERENCES \"User\" (\"Id\"), CONSTRAINT \"fk_Session_LastUsed\" FOREIGN KEY (\"LastUsed\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_Session_Id\" FOREIGN KEY (\"Id\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_Session_UserId\" FOREIGN KEY (\"UserId\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_Session_User\" FOREIGN KEY (\"User\") REFERENCES \"APIKey\" (\"Id\"), CONSTRAINT \"fk_Session_TokenHash\" FOREIGN KEY (\"TokenHash\") REFERENCES \"User\" (\"Id\"))").Error; err != nil {
		return err
	}
	return nil
}

func (m *Migration20250816103043) Down(db *gorm.DB) error {
	// Rollback operations in reverse order
	// Drop table Session
	if err := db.Exec("DROP TABLE IF EXISTS \"Session\"").Error; err != nil {
		return err
	}
	// Drop table SignedURL
	if err := db.Exec("DROP TABLE IF EXISTS \"SignedURL\"").Error; err != nil {
		return err
	}
	// Drop table File
	if err := db.Exec("DROP TABLE IF EXISTS \"File\"").Error; err != nil {
		return err
	}
	// Drop table NodeFileMetadata
	if err := db.Exec("DROP TABLE IF EXISTS \"NodeFileMetadata\"").Error; err != nil {
		return err
	}
	// Drop table APIKey
	if err := db.Exec("DROP TABLE IF EXISTS \"APIKey\"").Error; err != nil {
		return err
	}
	// Drop table Bucket
	if err := db.Exec("DROP TABLE IF EXISTS \"Bucket\"").Error; err != nil {
		return err
	}
	// Drop table User
	if err := db.Exec("DROP TABLE IF EXISTS \"User\"").Error; err != nil {
		return err
	}
	// Drop table SetupConfig
	if err := db.Exec("DROP TABLE IF EXISTS \"SetupConfig\"").Error; err != nil {
		return err
	}
	// Drop table StorageNode
	if err := db.Exec("DROP TABLE IF EXISTS \"StorageNode\"").Error; err != nil {
		return err
	}
	return nil
}
