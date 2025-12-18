# Default API URL (can be overridden with API_URL env var)
API_URL ?= http://localhost:8080

backfill: ## Run the backfill script to fill cache with popular repos
	@echo "Running backfill with API_URL=$(API_URL)"
	@go run ./cmd/fill-cache/backfill-repos.go



