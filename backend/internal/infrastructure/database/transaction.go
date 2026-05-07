package database

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Transaction provides a helper for managing Neo4j transactions
type Transaction struct {
	client *Neo4jClient
}

// NewTransaction creates a new transaction helper
func NewTransaction(client *Neo4jClient) *Transaction {
	return &Transaction{client: client}
}

// ReadTransaction executes a read transaction with automatic retry
func (t *Transaction) ReadTransaction(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := t.client.GetReadSession(ctx)
	defer session.Close(ctx)

	return session.ExecuteRead(ctx, work,
		neo4j.WithTxTimeout(30*1000), // 30 seconds
	)
}

// WriteTransaction executes a write transaction with automatic retry
func (t *Transaction) WriteTransaction(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := t.client.GetSession(ctx)
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, work,
		neo4j.WithTxTimeout(30*1000), // 30 seconds
	)
}

// BatchWrite executes multiple write operations in a single transaction
func (t *Transaction) BatchWrite(ctx context.Context, queries []CypherQuery) error {
	session := t.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, q := range queries {
			if _, err := tx.Run(ctx, q.Cypher, q.Params); err != nil {
				return nil, fmt.Errorf("batch query failed: %w", err)
			}
		}
		return nil, nil
	})

	return err
}

// CypherQuery represents a Cypher query with parameters
type CypherQuery struct {
	Cypher string
	Params map[string]interface{}
}

// QueryBuilder helps build Cypher queries
type QueryBuilder struct {
	cypher string
	params map[string]interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		params: make(map[string]interface{}),
	}
}

// Match adds a MATCH clause
func (qb *QueryBuilder) Match(pattern string) *QueryBuilder {
	qb.cypher += "MATCH " + pattern + " "
	return qb
}

// OptionalMatch adds an OPTIONAL MATCH clause
func (qb *QueryBuilder) OptionalMatch(pattern string) *QueryBuilder {
	qb.cypher += "OPTIONAL MATCH " + pattern + " "
	return qb
}

// Where adds a WHERE clause
func (qb *QueryBuilder) Where(condition string) *QueryBuilder {
	qb.cypher += "WHERE " + condition + " "
	return qb
}

// And adds an AND condition
func (qb *QueryBuilder) And(condition string) *QueryBuilder {
	qb.cypher += "AND " + condition + " "
	return qb
}

// Or adds an OR condition
func (qb *QueryBuilder) Or(condition string) *QueryBuilder {
	qb.cypher += "OR " + condition + " "
	return qb
}

// Return adds a RETURN clause
func (qb *QueryBuilder) Return(items string) *QueryBuilder {
	qb.cypher += "RETURN " + items + " "
	return qb
}

// Create adds a CREATE clause
func (qb *QueryBuilder) Create(pattern string) *QueryBuilder {
	qb.cypher += "CREATE " + pattern + " "
	return qb
}

// Merge adds a MERGE clause
func (qb *QueryBuilder) Merge(pattern string) *QueryBuilder {
	qb.cypher += "MERGE " + pattern + " "
	return qb
}

// OnCreate adds ON CREATE SET clause
func (qb *QueryBuilder) OnCreate(sets string) *QueryBuilder {
	qb.cypher += "ON CREATE SET " + sets + " "
	return qb
}

// OnMatch adds ON MATCH SET clause
func (qb *QueryBuilder) OnMatch(sets string) *QueryBuilder {
	qb.cypher += "ON MATCH SET " + sets + " "
	return qb
}

// Set adds a SET clause
func (qb *QueryBuilder) Set(assignments string) *QueryBuilder {
	qb.cypher += "SET " + assignments + " "
	return qb
}

// Delete adds a DELETE clause
func (qb *QueryBuilder) Delete(items string) *QueryBuilder {
	qb.cypher += "DELETE " + items + " "
	return qb
}

// DetachDelete adds a DETACH DELETE clause
func (qb *QueryBuilder) DetachDelete(items string) *QueryBuilder {
	qb.cypher += "DETACH DELETE " + items + " "
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(items string) *QueryBuilder {
	qb.cypher += "ORDER BY " + items + " "
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.cypher += fmt.Sprintf("LIMIT %d ", n)
	return qb
}

// Skip adds a SKIP clause
func (qb *QueryBuilder) Skip(n int) *QueryBuilder {
	qb.cypher += fmt.Sprintf("SKIP %d ", n)
	return qb
}

// With adds a WITH clause
func (qb *QueryBuilder) With(items string) *QueryBuilder {
	qb.cypher += "WITH " + items + " "
	return qb
}

// Unwind adds an UNWIND clause
func (qb *QueryBuilder) Unwind(expression string) *QueryBuilder {
	qb.cypher += "UNWIND " + expression + " "
	return qb
}

// Raw adds raw Cypher text
func (qb *QueryBuilder) Raw(cypher string) *QueryBuilder {
	qb.cypher += cypher + " "
	return qb
}

// Param adds a parameter
func (qb *QueryBuilder) Param(key string, value interface{}) *QueryBuilder {
	qb.params[key] = value
	return qb
}

// Params adds multiple parameters
func (qb *QueryBuilder) Params(params map[string]interface{}) *QueryBuilder {
	for k, v := range params {
		qb.params[k] = v
	}
	return qb
}

// Build returns the constructed query
func (qb *QueryBuilder) Build() CypherQuery {
	return CypherQuery{
		Cypher: qb.cypher,
		Params: qb.params,
	}
}

// String returns the Cypher query string
func (qb *QueryBuilder) String() string {
	return qb.cypher
}
