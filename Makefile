build:
	@echo "Building 'room' service..."
	@go build -o cmd/room/bin/room ./cmd/room
	@echo "Building 'mediator' service..."
	@go build -o cmd/mediator/bin/mediator ./cmd/mediator

dockerize:
	@echo "Building 'room' docker image..."
	@docker build -t gameon-a8-room/room:latest cmd/room
	@echo "Building 'mediator' docker image..."
	@docker build -t gameon-a8-room/mediator:latest cmd/mediator

start:
	@echo "Starting a8-room application..."
	@docker-compose up -d
	
stop:
	@echo "Stopping a8-room application..."
	@docker-compose stop 
	@docker-compose rm -f
