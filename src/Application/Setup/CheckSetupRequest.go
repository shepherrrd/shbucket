package setup

import (
	"context"
	"fmt"

	"shbucket/src/Infrastructure/Persistence"
)

type CheckSetupCommand struct{}

type CheckSetupResponse struct {
	IsSetup   bool   `json:"is_setup"`
	SetupType string `json:"setup_type,omitempty"`
	NodeName  string `json:"node_name,omitempty"`
	Message   string `json:"message"`
}

type CheckSetupRequestHandler struct {
	dbContext *persistence.AppDbContext
}

func NewCheckSetupRequestHandler(dbContext *persistence.AppDbContext) *CheckSetupRequestHandler {
	return &CheckSetupRequestHandler{
		dbContext: dbContext,
	}
}

func (h *CheckSetupRequestHandler) Handle(ctx context.Context, command *CheckSetupCommand) (*CheckSetupResponse, error) {
	config, err := h.dbContext.SetupConfigs.FirstOrDefault()
	if err != nil {
		return &CheckSetupResponse{
			IsSetup: false,
			Message: fmt.Sprintf("System not configured. Please complete initial setup %s.", err.Error()),
		}, nil
	}

	if config == nil {
			return &CheckSetupResponse{
				IsSetup: false,
				Message: "System not configured. Please complete initial setup.",
			}, nil
		}
	
 
	if !config.IsSetup {
		return &CheckSetupResponse{
			IsSetup: false,
			Message: "Setup in progress or incomplete. Please complete configuration.",
		}, nil
	}

	return &CheckSetupResponse{
		IsSetup:   true,
		SetupType: config.SetupType,
		NodeName:  config.NodeName,
		Message:   "System is configured and ready.",
	}, nil
}
