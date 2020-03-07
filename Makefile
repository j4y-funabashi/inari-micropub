build:
	go build -o bin/inari-web -v cmd/inari-web/main.go

test:
	./scripts/run_tests.sh

local:
	docker-compose -f docker-compose-local.yml down -v
	docker-compose -f docker-compose-local.yml up --build
