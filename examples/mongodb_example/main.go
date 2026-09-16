// Package main demonstrates complete MongoDB adapter capabilities:
// Dynamic $jsonSchema Validation, Flexible Document CRUD, Custom Commands & Compatibility Guards,
// Dynamic Query AST, Multi-Document Transactions, and Interactive Swagger UI.
//
// File: main.go
// Usage:
//   Serves as the entry point for the MongoDB example, executing the comprehensive
//   suite of demonstrations and optionally starting the HTTP Swagger UI server.
package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-mongodb"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"os"
)

// main is the CLI entrypoint for running the MongoDB demonstration suite.
//
// Purpose:
//   Invokes the complete set of MongoDB adapter demonstrations.
//
// Where it is used:
//   - When executing `go run main.go` or launching the binary.
//
// When can it be used:
//   - At application startup for demonstration or standalone server operation.
func main() {
	RunMongoDBExample()
}

// RunMongoDBExample runs the complete suite of MongoDB adapter demos and can launch the Swagger UI server.
//
// Purpose:
//   Initializes the MongoDB adapter and engine, executes CRUD, compatibility, pipeline,
//   raw command, and transaction demos, and starts the Swagger HTTP server if requested.
//
// Where it is used:
//   - Called by main() and automated integration test runners.
//
// When can it be used:
//   - When verifying end-to-end MongoDB engine integration or launching the example server.
func RunMongoDBExample() {
	ctx := context.Background()
	fmt.Println("=========================================================")
	fmt.Println("  MONGODB ADAPTER: COMPLETE CAPABILITIES DEMO")
	fmt.Println("=========================================================")

	uri := os.Getenv("MONGO_URI")
	mongoAdapter := mongodb.NewMongoAdapter(uri, "dev")
	engine := project.New(mongoAdapter)

	// 2. Run Dynamic Document CRUD Demo
	if err := RunMongoCRUDDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 3. Run Compatibility Guard (Function) Demo
	if err := RunMongoFunctionDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 4. Run Mongo Command & Pipeline (Procedure) Demo
	if err := RunMongoProcedureDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 5. Run Raw Command & Dynamic Query Demo
	if err := RunMongoRawQueryDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 6. Run Multi-Document Transaction Demo
	if err := RunMongoTransactionDemo(ctx, engine); err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	fmt.Println("\n=========================================================")
	fmt.Println("✔ ALL MONGODB DEMOS COMPLETED SUCCESSFULLY!")
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
