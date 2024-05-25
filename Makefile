OUTPUT_EX_SO_FOLDER := examples_shared_objects

run-coordinator:
	go run -race cmd/mrcoordinator.go ${INPUT_TEXT_FILE}

run-worker:
	go run -race cmd/mrworker.go ${SO}

compile-example-sos:
	echo "Compiling examples..."
	mkdir -p $(OUTPUT_EX_SO_FOLDER)
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/crash.go
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/early_exit.go
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/indexer.go
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/jobcount.go
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/mtiming.go
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/nocrash.go
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/rtiming.go
	go build -C $(OUTPUT_EX_SO_FOLDER) -race -buildmode=plugin ../examples/wc.go

clean:
	rm -rf $(OUTPUT_EX_SO_FOLDER)

