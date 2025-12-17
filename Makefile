.PHONY: dep-update
dep-update:
	go get -t -u=patch ./...
	go mod tidy

