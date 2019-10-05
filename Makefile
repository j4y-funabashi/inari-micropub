build:
	go build -o bin/inari-web -v cmd/inari-web/main.go

test:
	docker-compose down -v
	docker-compose up --build --exit-code-from tests

local:
	docker-compose -f docker-compose-local.yml down -v
	docker-compose -f docker-compose-local.yml up --build
