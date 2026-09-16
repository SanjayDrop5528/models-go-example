// Package main demonstrates complete Memory adapter capabilities:
// In-Memory Model Lifecycle, Diff Evolution, Dynamic CRUD, Operations, Query AST, Transactions, and Interactive Swagger UI.
//
// File: main.go
// Usage:
//   Serves as the entrypoint for the in-memory adapter demonstrations, running lifecycle demos
//   and optionally serving the Swagger UI REST server.
package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-memory"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"os"
)

// main is the CLI entrypoint for the in-memory demonstration suite.
//
// Purpose:
//   Calls RunMemoryExample to run the in-memory adapter test suite.
//
// Where it is used:
//   - Executed upon `go run main.go`.
//
// When can it be used:
//   - When executing the memory example executable.
func main() {
	RunMemoryExample()
}

// RunMemoryExample runs the complete suite of Memory adapter demos and can launch the Swagger UI server.
//
// Purpose:
//   Instantiates the in-memory engine, runs CRUD, functions, procedures, queries, and transactions,
//   and conditionally starts the HTTP Swagger documentation server.
//
// Where it is used:
//   - In main() and automated example test packages.
//
// When can it be used:
//   - When demonstrating or verifying in-memory storage features.
func RunMemoryExample() {
	ctx := context.Background()
	fmt.Println("=========================================================")
	fmt.Println("  MEMORY ADAPTER: COMPLETE CAPABILITIES DEMO")
	fmt.Println("=========================================================")

	// 1. Initialize Memory Adapter and Engine directly (zero config needed)
	memAdapter := memory.NewMemoryAdapter()
	engine := project.New(memAdapter)

	// 2. Run Dynamic CRUD Demo
	if err := RunMemoryCRUDDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 3. Run Custom Function Demo
	if err := RunMemoryFunctionDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 4. Run Stored Procedure Demo
	if err := RunMemoryProcedureDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 5. Run Raw Query & Dynamic AST Demo
	if err := RunMemoryRawQueryDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 6. Run Atomic Transaction Demo
	if err := RunMemoryTransactionDemo(ctx, engine); err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	fmt.Println("\n=========================================================")
	fmt.Println("✔ ALL MEMORY ADAPTER DEMOS COMPLETED SUCCESSFULLY!")
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
