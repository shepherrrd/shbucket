package services

import (
	"shbucket/src/Application/File"
	"shbucket/src/Infrastructure/Config"
	"shbucket/src/Infrastructure/Data/Entities"
	"shbucket/src/Infrastructure/Persistence"
)

// SignatureValidationService provides centralized signature validation and management
type SignatureValidationService struct {
	dbContext *persistence.AppDbContext
	settings  *config.Settings
	signedURLHandler *file.GenerateSignedURLRequestHandler
}

// NewSignatureValidationService creates a new instance of SignatureValidationService
func NewSignatureValidationService(dbContext *persistence.AppDbContext) *SignatureValidationService {
	return &SignatureValidationService{
		dbContext: dbContext,
		settings:  config.GetSettings(),
		signedURLHandler: file.NewGenerateSignedURLRequestHandler(dbContext),
	}
}

// ValidateAndConsumeSignature is deprecated - use ValidateSignatureOnly and MarkSignatureAsUsed separately
// This was overcomplicated, the simple approach in ServeFile is better
func (s *SignatureValidationService) ValidateAndConsumeSignature(signature string) (*entities.SignedURL, *entities.File, *entities.Bucket, error) {
	// First validate the signature
	signedURL, err := s.signedURLHandler.ValidateSignedURL(signature)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get file and bucket information
	file, bucket, err := s.signedURLHandler.GetFileInfoFromSignature(signature)
	if err != nil {
		return nil, nil, nil, err
	}

	// Mark as used if it's a single-use URL
	if signedURL.SingleUse {
		if err := s.signedURLHandler.MarkSignatureAsUsed(signature); err != nil {
			return nil, nil, nil, err
		}
	}

	return signedURL, file, bucket, nil
}

// ValidateSignatureOnly validates a signature without marking it as used
// Use this for checking validity without consuming the signature
func (s *SignatureValidationService) ValidateSignatureOnly(signature string) (*entities.SignedURL, error) {
	return s.signedURLHandler.ValidateSignedURL(signature)
}

// GetFileInfoFromSignature returns file and bucket information from a signature
func (s *SignatureValidationService) GetFileInfoFromSignature(signature string) (*entities.File, *entities.Bucket, error) {
	return s.signedURLHandler.GetFileInfoFromSignature(signature)
}

// MarkSignatureAsUsed manually marks a signature as used (for single-use URLs)
func (s *SignatureValidationService) MarkSignatureAsUsed(signature string) error {
	return s.signedURLHandler.MarkSignatureAsUsed(signature)
}