// Package main demonstrates complete MySQL adapter capabilities:
// Dynamic Model Schema Migrations, Dynamic CRUD, Stored Functions, Stored Procedures,
// Raw SQL Query AST, Atomic Multi-Model Transactions, and Interactive Swagger UI.
package main

import (
	"context"
	"fmt"
	"github.com/SanjayDrop5528/models-go-mysql"
	"github.com/SanjayDrop5528/models-go-engine/project"
	"os"
)

func main() {
	RunMySQLExample()
}

// RunMySQLExample runs the complete suite of MySQL adapter demos and can launch the Swagger UI server.
func RunMySQLExample() {
	ctx := context.Background()
	fmt.Println("=========================================================")
	fmt.Println("  MYSQL ADAPTER: COMPLETE CAPABILITIES DEMO")
	fmt.Println("=========================================================")

	// 1. Initialize MySQL Adapter and Engine directly (zero config needed)
	dsn := os.Getenv("MYSQL_DSN")
	mysqlAdapter := mysql.NewMySQLAdapter(dsn)
	engine := project.New(mysqlAdapter)

	// 2. Run Dynamic CRUD Demo
	if err := RunMySQLCRUDDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 3. Run Custom Function Demo
	if err := RunMySQLFunctionDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 4. Run Stored Procedure Demo
	if err := RunMySQLProcedureDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 5. Run Raw SQL & Query AST Demo
	if err := RunMySQLRawQueryDemo(ctx, engine); err != nil {
		panic(err)
	}

	// 6. Run Atomic Transaction Demo
	if err := RunMySQLTransactionDemo(ctx, engine); err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	fmt.Println("\n=========================================================")
	fmt.Println("✔ ALL MYSQL DEMOS COMPLETED SUCCESSFULLY!")
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
