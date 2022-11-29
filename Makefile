test-all:
	go test ./... -coverprofile=coverage.out

cover:
	go tool cover -html=coverage.out

clean:
	rm -f coverage.out