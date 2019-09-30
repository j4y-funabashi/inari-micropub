build:
	go build -o bin/inari-web -v cmd/inari-web/main.go

test:
	docker-compose down -v
	docker-compose up --build --exit-code-from tests
