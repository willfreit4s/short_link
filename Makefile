DB_URL ?= postgres://postgres:postgres@localhost:5432/short_link?sslmode=disable

migUP:
	migrate -path=sql/migrations -database "$(DB_URL)" -verbose up

migDOWN:
	migrate -path=sql/migrations -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

run:
	go run cmd/server/main.go

docker-up:
	docker compose -f docker/docker-compose.yaml up --build

docker-up-d:
	docker compose -f docker/docker-compose.yaml up --build -d

docker-down:
	docker compose -f docker/docker-compose.yaml down

# Reset database
docker-down-v:
	docker compose -f docker/docker-compose.yaml down -v

docker-logs:
	docker compose -f docker/docker-compose.yaml logs -f

docker-logs-api:
	docker compose -f docker/docker-compose.yaml logs -f api

## Runs only unit tests (which are fast and have no external dependencies)
test-unit:
	go test ./... -short -race -cover

## Runs only integration tests (uses testcontainers, requires Docker)
test-integration:
	go test ./test/... -tags=integration -race -v

## Runs the entire suite of tests (unit and integration)
test-all: test-unit test-integration

## Generate a coverage report (unit tests only)
test-coverage:
	go test ./... -short -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Kubernetes

K8S_NAMESPACE=short-link
KIND_CLUSTER=short-link

k8s-create:
	kind create cluster --config k8s/kind-config.yaml

k8s-delete:
	kind delete cluster --name $(KIND_CLUSTER)

k8s-status:
	kubectl get nodes

k8s-pods:
	kubectl get pods -n $(K8S_NAMESPACE)

k8s-apply-namespace:
	kubectl apply -f k8s/namespace.yaml

k8s-build:
	docker build -t short-link:local .

k8s-load-image:
	kind load docker-image short-link:local --name $(KIND_CLUSTER)

k8s-deploy:
	kubectl apply -f k8s/api/deployment.yaml

k8s-deployment:
	kubectl get deployment -n $(K8S_NAMESPACE)

k8s-pods-wide:
	kubectl get pods -n $(K8S_NAMESPACE) -o wide

.PHONY: \
	migUP \
	migDOWN \
	sqlc \
	run \
	docker-up \
	docker-up-d \
	docker-down \
	docker-down-v \
	docker-logs \
	docker-logs-api \
	test-unit \
	test-integration \
	test-all \
	test-coverage \
	k8s-create \
	k8s-delete \
	k8s-status \
	k8s-pods \
	k8s-apply-namespace \
	k8s-build \
	k8s-load-image \
	k8s-deploy \
	k8s-deployment \
	k8s-pods-wide