package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Neo4j      Neo4jConfig
	Kubernetes KubernetesConfig
	GraphQL    GraphQLConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	Host        string
	Environment    string
	Debug          bool
	AllowedOrigins []string
}

// Neo4jConfig holds Neo4j database configuration
type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	KubeconfigPath string
	InCluster      bool
}

// GraphQLConfig holds GraphQL configuration
type GraphQLConfig struct {
	PlaygroundEnabled    bool
	IntrospectionEnabled bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for development)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Environment:    getEnv("APP_ENV", "development"),
			Debug:          getEnvBool("DEBUG", true),
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		},
		Neo4j: Neo4jConfig{
			URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
			Username: getEnv("NEO4J_USERNAME", "neo4j"),
			Password: getEnv("NEO4J_PASSWORD", "password"),
			Database: getEnv("NEO4J_DATABASE", "neo4j"),
		},
		Kubernetes: KubernetesConfig{
			KubeconfigPath: getEnv("KUBECONFIG", ""),
			InCluster:      getEnvBool("IN_CLUSTER", false),
		},
		GraphQL: GraphQLConfig{
			PlaygroundEnabled:    getEnvBool("GRAPHQL_PLAYGROUND", true),
			IntrospectionEnabled: getEnvBool("GRAPHQL_INTROSPECTION", true),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Neo4j.URI == "" {
		return fmt.Errorf("neo4j URI is required")
	}
	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolVal
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intVal
	}
	return defaultValue
}

// getEnvSlice gets a slice of strings from a comma-separated environment variable
func getEnvSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		if value == "" {
			return []string{}
		}
		parts := strings.Split(value, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	return defaultValue
}
