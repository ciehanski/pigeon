run: # Builds and runs the project
	docker-compose up -d
stop: # Stops the project
	docker-compose down -v --remove-orphans
restart: stop run
reset: # Refreshes base pigeon docker image and restarts the project
	docker rmi -f pigeon_pigeon:latest
	docker-compose up -d
build: # Builds the pigeon artifact
	go get -u -a -v -x github.com/ipsn/go-libtor
	go mod download
	cd cmd/pigeon && CGO_ENABLED=1 go build -a -installsuffix cgo -ldflags '-s' -o pigeon .
logs: # Prints docker-compose logs
	docker-compose logs -f --tail 100 pigeon
exec: # Open a bash shell into the docker container
	docker exec -it pigeon bash
lint: # Will lint the project
	golint ./...
	go fmt ./...
test: # Will run tests on the project
	go test -v -race -bench=. -cpu=1,2,4 ./... && \
	go vet ./...

.PHONY: run stop restart reset build logs exec
