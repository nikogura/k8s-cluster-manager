.PHONY: lint

lint:
	namedreturns ./...
	golangci-lint run
