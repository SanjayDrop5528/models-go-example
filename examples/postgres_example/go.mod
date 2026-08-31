module models-go/examples/postgres_example

go 1.26.2

require (
	models-go/adapters/postgres v0.0.0
	models-go/engine v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace (
	models-go/adapters/postgres => ../../adapters/postgres
	models-go/engine => ../../engine
)
