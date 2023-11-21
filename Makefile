push:
	go mod download && go mod vendor && git add . && git commit -m '$(m)'

build:
	go build -o bin/$(name) cmd/$(name)/main.go

test:
	export RUN_ENVIRONMENT=test
	go test -race $(go list ./...)