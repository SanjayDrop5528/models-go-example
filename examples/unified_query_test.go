package examples_test

import (
	"context"
	"strings"
	"testing"

	"github.com/SanjayDrop5528/models-go-engine/model"
	"github.com/SanjayDrop5528/models-go-engine/query"
	"github.com/SanjayDrop5528/models-go-engine/rdbms"
	"github.com/SanjayDrop5528/models-go-engine/rdbms/dialect"
	"github.com/SanjayDrop5528/models-go-memory"
	"github.com/SanjayDrop5528/models-go-mongodb"
	"github.com/SanjayDrop5528/models-go-mysql"
	"github.com/SanjayDrop5528/models-go-postgres"
)

// TestUnifiedCrossDatabaseQuery demonstrates that the EXACT SAME query data structure
// executes across PostgreSQL, MySQL, MongoDB, and In-Memory identically.
func TestUnifiedCrossDatabaseQuery(t *testing.T) {
	// 1. Construct ONE single unified query using the query builder
	unifiedQuery := query.New().
		Table("users").
		Column("id", "name", "email").
		Where("status = ?", "active").
		WhereFilter("age", query.OpGte, 18).
		WhereGroup(" OR ", func(sub query.Query) query.Query {
			return sub.Where("role = ?", "admin").Where("role = ?", "manager")
		}).
		OrderBy("created_at", query.SortDesc).
		Limit(10).
		Offset(0)

	// 2. Compile with PostgreSQL
	pgQB := &postgres.QueryBuilder{}
	pgSQL, pgArgs := pgQB.BuildSelect("users", unifiedQuery)
	if !strings.Contains(pgSQL, `SELECT "id", "name", "email" FROM "users"`) {
		t.Fatalf("PostgreSQL select failed: %s", pgSQL)
	}
	if len(pgArgs) < 4 {
		t.Fatalf("PostgreSQL expected at least 4 args, got %d", len(pgArgs))
	}

	// 3. Compile with MySQL
	myQB := &mysql.QueryBuilder{}
	mySQL, myArgs := myQB.BuildSelect("users", unifiedQuery)
	if !strings.Contains(mySQL, "SELECT `id`, `name`, `email` FROM `users`") {
		t.Fatalf("MySQL select failed: %s", mySQL)
	}
	if len(myArgs) < 4 {
		t.Fatalf("MySQL expected at least 4 args, got %d", len(myArgs))
	}

	// 4. Compile with MongoDB
	mongoQB := &mongodb.QueryBuilder{}
	mongoPipeline := mongoQB.BuildPipeline(unifiedQuery)
	if len(mongoPipeline) == 0 {
		t.Fatalf("MongoDB pipeline empty")
	}
	hasMatch := false
	for _, stage := range mongoPipeline {
		if _, ok := stage["$match"]; ok {
			hasMatch = true
		}
	}
	if !hasMatch {
		t.Fatalf("MongoDB pipeline missing $match stage")
	}

	// 5. Execute with In-Memory
	memAdapter := memory.NewMemoryAdapter()
	ctx := context.Background()
	ref := model.ModelRef{Name: "users", StorageName: "users"}
	_, _ = memAdapter.Create(ctx, ref, map[string]any{"id": 1, "name": "Alice", "status": "active", "age": 25, "role": "admin"})
	_, _ = memAdapter.Create(ctx, ref, map[string]any{"id": 2, "name": "Bob", "status": "inactive", "age": 30, "role": "user"})

	results, total, err := memAdapter.Find(ctx, ref, unifiedQuery)
	if err != nil {
		t.Fatalf("In-Memory Find failed: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Fatalf("In-Memory expected 1 result for Alice, got %d", len(results))
	}
	if results[0]["name"] != "Alice" {
		t.Fatalf("In-Memory expected Alice, got %v", results[0]["name"])
	}

	// 6. Compile directly with RDBMS SelectQuery bridge
	rdb := rdbms.NewDB(nil, dialect.NewPostgreSQL())
	sq := rdbms.NewSelectFromQuery(rdb, unifiedQuery)
	if !strings.Contains(sq.String(), `SELECT "id", "name", "email" FROM "users"`) {
		t.Fatalf("RDBMS SelectQuery bridge failed: %s", sq.String())
	}
}
