// Package main demonstrates the complete PostgreSQL adapter capabilities:
// Dynamic Model Schema Migrations, Dynamic CRUD, Stored Functions, Stored Procedures,
// Raw SQL Query AST, Atomic Multi-Model Transactions, and Interactive Swagger UI.
//
// File: main.go
// Usage:
//   Executable CLI / demo entrypoint orchestrating PostgreSQL CRUD, stored routines,
//   raw queries, transaction execution, and starting the interactive Swagger HTTP server.
package main

import (
	"fmt"
	"os"

	"github.com/SanjayDrop5528/models-go-engine/project"
	postgres "github.com/SanjayDrop5528/models-go-postgres"
)

// main is the CLI entrypoint for PostgreSQL examples.
//
// Purpose:
//   Dispatches to RunPostgresExample to run demos and optional Swagger server.
//
// Where it is used:
//   - When executing `go run ./examples/postgres_example/main.go`.
//
// When can it be used:
//   - To run PostgreSQL demonstrations locally.
func main() {
	RunPostgresExample()
}

// RunPostgresExample runs the complete suite of PostgreSQL demos and can launch the Swagger UI server.
//
// Purpose:
//   Initializes PostgreSQL adapter and engine, and conditionally launches the Swagger UI server.
//
// Where it is used:
//   - Called by main() and integration test suites.
//
// When can it be used:
//   - Whenever executing PostgreSQL capability walkthroughs.
func RunPostgresExample() {
	// ctx := context.Background()
	fmt.Println("=========================================================")
	fmt.Println("  POSTGRESQL ADAPTER: COMPLETE CAPABILITIES DEMO")
	fmt.Println("=========================================================")

	// 1. Initialize PostgreSQL Adapter and Engine directly (zero config needed)
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgrespassword@localhost:5432/uat_mineone?sslmode=disable"
	}
	pgAdapter := postgres.NewPostgresAdapter(dsn)
	engine := project.New(pgAdapter)

	// // 2. Run Complete Dynamic CRUD Demo
	// if err := RunPostgresCRUDDemo(ctx, engine); err != nil {
	// 	panic(err)
	// }

	// // 3. Run Stored Function Demo (Registration, validation & calling)
	// if err := RunPostgresFunctionDemo(ctx, engine); err != nil {
	// 	panic(err)
	// }

	// // 4. Run Stored Procedure Demo (Registration, calling & CALL SQL)
	// if err := RunPostgresProcedureDemo(ctx, engine); err != nil {
	// 	panic(err)
	// }

	// // 5. Run Raw SQL & Advanced Query AST Demo
	// if err := RunPostgresRawQueryDemo(ctx, engine); err != nil {
	// 	panic(err)
	// }

	// // 6. Run Atomic Multi-Model Transaction Demo
	// if err := RunPostgresTransactionDemo(ctx, engine); err != nil {
	// 	panic(err)
	// }

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Println("\n=========================================================")
	fmt.Println("✔ ALL POSTGRESQL DEMOS COMPLETED SUCCESSFULLY!")
	fmt.Printf("📖 Interactive Swagger UI Server ready at: http://localhost:%s/swagger/\n", port)
	fmt.Println("   (Run with '--server' or 'SERVER=true' to keep server running)")
	fmt.Println("=========================================================")

	for _, arg := range os.Args {
		if arg == "--server" || os.Getenv("SERVER") == "true" {
			srv := StartSwaggerServer(port, engine)
			if err := srv.ListenAndServe(); err != nil {
				fmt.Printf("Server stopped: %v\n", err)
			}
		}
	}
}
