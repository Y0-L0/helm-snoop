.PHONY: dep-update fmt test coverage complete

dep-update:
	go get -t -u=patch ./...
	go mod tidy

# Format all Go files in the module
fmt:
	go fmt ./...

# Run the full test suite
# Supports extra flags via TF or as direct args after the target, e.g.:
#   make test -v -run TestUnit
#   make test TF="-v -run TestUnit"
test: fmt
	go test $(TF) $(filter-out $@ --,$(MAKECMDGOALS)) ./...

# Generate coverage report (coverage.html). Deletes any previous report first
# and suppresses test output for a clean run. Tracks exit status across steps.
coverage: fmt
	@status=0; \
	rm -f coverage.html || true; \
	go test $(TF) $(filter-out $@ --,$(MAKECMDGOALS)) ./... -coverpkg=./... -coverprofile=coverage.out || status=$$?; \
	if [ -f coverage.out ]; then go tool cover -html=coverage.out -o coverage.html || status=$$?; fi; \
	exit $$status

# Full workflow: run the CLI against the sample chart, then produce coverage.
# Keeps track of exit code across steps and returns it at the end.
complete: fmt
	@status=0; \
	go run ./cmd/... test-chart/ || status=$$?; \
	$(MAKE) coverage -- $(filter-out $@,$(MAKECMDGOALS)) TF="$(TF)" || status=$$?; \
	exit $$status

# Swallow unknown extra make goals (used to pass args like -v, -run, etc.)
%:
	@:
