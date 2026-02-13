.PHONY: build dep-update fmt test coverage complete test-pre-commit-hook

# Build binary for the local OS/arch using GoReleaser (snapshot mode)
build:
	goreleaser build --clean --snapshot --single-target

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
	go run ./cmd/helm-snoop/... testdata/test-chart/ || status=$$?; \
	$(MAKE) coverage -- $(filter-out $@,$(MAKECMDGOALS)) TF="$(TF)" || status=$$?; \
	exit $$status

# Test the pre-commit hook definition against the sample chart
test-pre-commit-hook:
	prek try-repo . helm-snoop helm-snoop-docker --files testdata/test-chart/Chart.yaml testdata/test-chart/values.yaml testdata/test-chart/templates/configmap.yaml testdata/test-chart/templates/deployment.yaml testdata/test-chart/templates/_defs.tpl testdata/test-chart/templates/_names.tpl testdata/test-chart/templates/_secrets.tpl

# Swallow unknown extra make goals (used to pass args like -v, -run, etc.)
%:
	@:
