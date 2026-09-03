.PHONY: dev prod down logs test lint tidy network

# Create the external shared network if it doesn't already exist.
network:
	docker network inspect app-bridge >/dev/null 2>&1 || docker network create app-bridge

dev: network
	@test -f .env || (echo "Missing .env — copy .env.example to .env first" && exit 1)
	docker compose --profile dev up --build

prod: network
	@test -f .env || (echo "Missing .env — copy .env.example to .env first" && exit 1)
	docker compose --profile prod up --build -d

down:
	docker compose --profile dev --profile prod down

logs:
	docker compose logs -f

test:
	cd backend && go test ./... -cover

lint:
	cd backend && go vet ./... && gofmt -l .

tidy:
	cd backend && go mod tidy
