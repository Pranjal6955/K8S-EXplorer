package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/K8S-Graph-Explorer/backend/internal/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jClient wraps the Neo4j driver with singleton pattern
type Neo4jClient struct {
	driver   neo4j.DriverWithContext
	database string
	config   config.Neo4jConfig
}

var (
	instance *Neo4jClient
	once     sync.Once
	initErr  error
)

// GetInstance returns the singleton Neo4j client instance
// Must call Initialize() first or this will return nil
func GetInstance() *Neo4jClient {
	return instance
}

// Initialize creates the singleton Neo4j client instance
// This should be called once at application startup
func Initialize(cfg config.Neo4jConfig) error {
	once.Do(func() {
		instance, initErr = newNeo4jClient(cfg)
	})
	return initErr
}

// MustInitialize creates the singleton and panics on error
func MustInitialize(cfg config.Neo4jConfig) *Neo4jClient {
	if err := Initialize(cfg); err != nil {
		panic(fmt.Sprintf("failed to initialize Neo4j: %v", err))
	}
	return instance
}

// NewNeo4jClient creates a new Neo4j client (non-singleton)
// Use this for testing or when you need multiple connections
func NewNeo4jClient(cfg config.Neo4jConfig) (*Neo4jClient, error) {
	return newNeo4jClient(cfg)
}

// newNeo4jClient is the internal constructor
func newNeo4jClient(cfg config.Neo4jConfig) (*Neo4jClient, error) {
	// Create driver with configuration
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.Username, cfg.Password, ""),
		func(config *neo4j.Config) {
			config.MaxConnectionPoolSize = 50
			config.MaxConnectionLifetime = 1 * time.Hour
			config.ConnectionAcquisitionTimeout = 1 * time.Minute
			config.SocketConnectTimeout = 30 * time.Second
			config.SocketKeepalive = true
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	return &Neo4jClient{
		driver:   driver,
		database: cfg.Database,
		config:   cfg,
	}, nil
}

// Close closes the Neo4j driver connection
func (c *Neo4jClient) Close() error {
	if c.driver != nil {
		return c.driver.Close(context.Background())
	}
	return nil
}

// CloseInstance closes the singleton instance
func CloseInstance() error {
	if instance != nil {
		return instance.Close()
	}
	return nil
}

// GetDriver returns the underlying Neo4j driver
func (c *Neo4jClient) GetDriver() neo4j.DriverWithContext {
	return c.driver
}

// GetSession creates a new session for the database
func (c *Neo4jClient) GetSession(ctx context.Context) neo4j.SessionWithContext {
	return c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
		AccessMode:   neo4j.AccessModeWrite,
	})
}

// GetReadSession creates a new read-only session
func (c *Neo4jClient) GetReadSession(ctx context.Context) neo4j.SessionWithContext {
	return c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
		AccessMode:   neo4j.AccessModeRead,
	})
}

// ExecuteRead executes a read transaction
func (c *Neo4jClient) ExecuteRead(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := c.GetReadSession(ctx)
	defer session.Close(ctx)

	return session.ExecuteRead(ctx, work)
}

// ExecuteWrite executes a write transaction
func (c *Neo4jClient) ExecuteWrite(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := c.GetSession(ctx)
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, work)
}

// RunQuery executes a Cypher query and returns the result records
func (c *Neo4jClient) RunQuery(ctx context.Context, cypher string, params map[string]interface{}) ([]map[string]interface{}, error) {
	session := c.GetReadSession(ctx)
	defer session.Close(ctx)

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect results: %w", err)
	}

	// Convert records to maps
	var output []map[string]interface{}
	for _, record := range records {
		recordMap := make(map[string]interface{})
		for _, key := range record.Keys {
			value, _ := record.Get(key)
			recordMap[key] = c.convertNeo4jValue(value)
		}
		output = append(output, recordMap)
	}

	return output, nil
}

// RunWrite executes a write Cypher query
func (c *Neo4jClient) RunWrite(ctx context.Context, cypher string, params map[string]interface{}) (*neo4j.ResultSummary, error) {
	session := c.GetSession(ctx)
	defer session.Close(ctx)

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("write query failed: %w", err)
	}

	summary, err := result.Consume(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to consume result: %w", err)
	}

	return &summary, nil
}

// Ping checks if the database is reachable
func (c *Neo4jClient) Ping(ctx context.Context) error {
	return c.driver.VerifyConnectivity(ctx)
}

// Health returns health status of the connection
func (c *Neo4jClient) Health(ctx context.Context) (bool, string) {
	if err := c.Ping(ctx); err != nil {
		return false, fmt.Sprintf("Neo4j connection failed: %v", err)
	}
	return true, "Neo4j connection healthy"
}

// convertNeo4jValue converts Neo4j types to Go types
func (c *Neo4jClient) convertNeo4jValue(value interface{}) interface{} {
	switch v := value.(type) {
	case neo4j.Node:
		return map[string]interface{}{
			"id":         v.ElementId,
			"labels":     v.Labels,
			"properties": v.Props,
		}
	case neo4j.Relationship:
		return map[string]interface{}{
			"id":         v.ElementId,
			"type":       v.Type,
			"startNode":  v.StartElementId,
			"endNode":    v.EndElementId,
			"properties": v.Props,
		}
	case neo4j.Path:
		nodes := make([]interface{}, len(v.Nodes))
		for i, node := range v.Nodes {
			nodes[i] = c.convertNeo4jValue(node)
		}
		rels := make([]interface{}, len(v.Relationships))
		for i, rel := range v.Relationships {
			rels[i] = c.convertNeo4jValue(rel)
		}
		return map[string]interface{}{
			"nodes":         nodes,
			"relationships": rels,
		}
	default:
		return v
	}
}

// Stats returns connection pool statistics
func (c *Neo4jClient) Stats() map[string]interface{} {
	return map[string]interface{}{
		"database": c.database,
		"uri":      c.config.URI,
	}
}
