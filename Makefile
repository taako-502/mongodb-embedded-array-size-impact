.PHONY: db db-stop benchmark test-benchmark

db:
	docker run -d --name go_vs_aggregation_pipeline  -p 27017:27017 mongo

db-stop:
	docker stop go_vs_aggregation_pipeline  || true
	docker rm go_vs_aggregation_pipeline  || true

benchmark: db-stop
	@$(MAKE) db
	@echo "Running benchmark tests..."
	@go test -bench . -benchmem -timeout=60m | tee benchmark_results.txt || true
	@$(MAKE) db-stop
