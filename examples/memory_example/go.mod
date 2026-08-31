module models-go/examples/memory_example

go 1.26.2

require (
	models-go/adapters/memory v0.0.0
	models-go/engine v0.0.0
)

replace (
	models-go/adapters/memory => ../../adapters/memory
	models-go/engine => ../../engine
)
