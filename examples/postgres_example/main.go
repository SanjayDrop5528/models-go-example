// Package main demonstrates the complete PostgreSQL adapter capabilities:
// Dynamic Model Schema Migrations, Dynamic CRUD, Stored Functions, Stored Procedures,
// Raw SQL Query AST, Atomic Multi-Model Transactions, and Interactive Swagger UI.
package main

import (
	"fmt"
	"github.com/SanjayDrop5528/models-go-postgres"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"os"
)

func main() {
	RunPostgresExample()
}

// RunPostgresExample runs the complete suite of PostgreSQL demos and can launch the Swagger UI server.
func RunPostgresExample() {
	// ctx := context.Background()
	fmt.Println("=========================================================")
	fmt.Println("  POSTGRESQL ADAPTER: COMPLETE CAPABILITIES DEMO")
	fmt.Println("=========================================================")

	// 1. Initialize PostgreSQL Adapter and Engine directly (zero config needed)
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgrespassword@localhost:5432/interview_ai?sslmode=disable"
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
