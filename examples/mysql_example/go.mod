module models-go/examples/mysql_example

go 1.26.2

require (
	models-go/adapters/mysql v0.0.0
	models-go/engine v0.0.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
)

replace (
	models-go/adapters/mysql => ../../adapters/mysql
	models-go/engine => ../../engine
)
