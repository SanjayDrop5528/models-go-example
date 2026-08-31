// Package main demonstrates complete MongoDB adapter capabilities:
// Dynamic $jsonSchema Validation, Flexible Document CRUD, Custom Commands & Compatibility Guards,
// Dynamic Query AST, Multi-Document Transactions, and Interactive Swagger UI.
package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-mongodb"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"os"
)

func main() {
	RunMongoDBExample()
}

// RunMongoDBExample runs the complete suite of MongoDB adapter demos and can launch the Swagger UI server.
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
