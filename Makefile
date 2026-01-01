.PHONY: build clean zip deploy

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go

clean:
	rm -f bootstrap bootstrap.zip

zip: build
	zip bootstrap.zip bootstrap

deploy: zip
	@echo "Deployment is handled by GitHub Actions"
	@echo "To deploy manually, use: aws lambda update-function-code --function-name <function-name> --zip-file fileb://bootstrap.zip"

test:
	go test -v ./...

.DEFAULT_GOAL := build
