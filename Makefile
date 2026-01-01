.PHONY: build clean test docker-build docker-run

build:
	go build -o leaderboard-computer main.go

clean:
	rm -f leaderboard-computer

docker-build:
	docker build -t leaderboard-computer:latest .

docker-run: docker-build
	docker run --rm \
		-e DB_HOST=${DB_HOST} \
		-e DB_PORT=${DB_PORT} \
		-e DB_USER=${DB_USER} \
		-e DB_PASSWORD=${DB_PASSWORD} \
		-e DB_NAME=${DB_NAME} \
		leaderboard-computer:latest

test:
	go test -v ./...

.DEFAULT_GOAL := build
