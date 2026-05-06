package graph

import (
	"github.com/K8S-Graph-Explorer/backend/internal/interface/http/middleware/auth"
	"github.com/K8S-Graph-Explorer/backend/internal/repositories/neo4j"
	"github.com/K8S-Graph-Explorer/backend/internal/services"
	"github.com/K8S-Graph-Explorer/backend/internal/services/converter"
)

// Resolver is the root resolver with all dependencies
type Resolver struct {
	NodeRepo       *neo4j.NodeRepository
	EdgeRepo       *neo4j.EdgeRepository
	GraphConverter *converter.GraphConverter
	JWTService     *auth.JWTService
	SyncService    *services.SyncService
}

// NewResolver creates a new resolver with dependencies
func NewResolver(
	nodeRepo *neo4j.NodeRepository,
	edgeRepo *neo4j.EdgeRepository,
	graphConverter *converter.GraphConverter,
	jwtService *auth.JWTService,
	syncService *services.SyncService,
) *Resolver {
	return &Resolver{
		NodeRepo:       nodeRepo,
		EdgeRepo:       edgeRepo,
		GraphConverter: graphConverter,
		JWTService:     jwtService,
		SyncService:    syncService,
	}
}
