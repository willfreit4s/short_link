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
	k8s-create \
	k8s-delete \
	k8s-status \
	k8s-pods \
	k8s-apply-namespace