test:
	./scripts/run_tests.sh

build:
	go build -o bin/inari-web -v .
	go build -o bin/inari-replay -v cmd/inari-replay/main.go
