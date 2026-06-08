package registry

import (
	"context"
	"fmt"
	"io"
)

// RegistryArtifactProvider bridges the registry Service to the evaluator core ArtifactProvider interface.
type RegistryArtifactProvider struct {
	service *Service
}

// NewRegistryArtifactProvider creates a new RegistryArtifactProvider.
func NewRegistryArtifactProvider(service *Service) (*RegistryArtifactProvider, error) {
	if service == nil {
		return nil, fmt.Errorf("registry service is required")
	}
	return &RegistryArtifactProvider{service: service}, nil
}

// OpenPrivateArtifact retrieves the private evaluation artifact read stream for a given exercise version.
func (p *RegistryArtifactProvider) OpenPrivateArtifact(ctx context.Context, orgID, labID, version string) (io.ReadCloser, error) {
	exerciseVersion, err := p.service.GetExerciseVersion(ctx, orgID, labID, version)
	if err != nil {
		return nil, fmt.Errorf("get exercise version: %w", err)
	}

	handle, err := p.service.OpenArtifact(ctx, exerciseVersion.PrivateArtifactSHA)
	if err != nil {
		return nil, fmt.Errorf("open private artifact stream: %w", err)
	}

	return handle.File, nil
}
