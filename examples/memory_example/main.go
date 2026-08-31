// Package main demonstrates complete Memory adapter capabilities:
// In-Memory Model Lifecycle, Diff Evolution, Dynamic CRUD, Operations, Query AST, Transactions, and Interactive Swagger UI.
package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-memory"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"os"
)

func main() {
	RunMemoryExample()
}

// RunMemoryExample runs the complete suite of Memory adapter demos and can launch the Swagger UI server.
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
