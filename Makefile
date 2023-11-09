push:
	go mod download && go mod vendor && git add . && git commit -m '$(m)'

build:
	#docker build -f ./cmd/api/Dockerfile -t '$(tag)' .
	docker build -t registry.gitlab.com/mrqter/echo-admin:$(tag) -f cmd/api/Dockerfile .
	docker push registry.gitlab.com/mrqter/echo-admin:$(tag)

test:
	export RUN_ENVIRONMENT=test
	go test -race $(go list ./...)