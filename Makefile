.PHONY: all
all: test validate

.PHONY: test
test:
	GO_ENV=testing go test -v -race -cover ./...

.PHONY: validate
validate:
	./validate.sh
