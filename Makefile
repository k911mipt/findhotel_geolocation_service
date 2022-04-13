ENV?=local.env
include $(ENV)

###############################################################################
#                                   Docker                                    #
###############################################################################

docker.postgres.start:
	docker-compose up -d db

docker.api.start:
	docker-compose build
	docker-compose up -d api

docker.all.stop:
	docker-compose down

docker.all.restart:
	docker-compose down
	docker-compose build
	docker-compose up -d

docker.all.start:
	docker-compose build
	docker-compose up -d

###############################################################################
#                                   Import                                    #
###############################################################################

import.build:
	go build -a  -o ./bin/import ./cmd/import

import.run: import.build
	./bin/import

###############################################################################
#                                     Api                                     #
###############################################################################

api.build:
	go build -a -o ./bin/api ./cmd/api

api.run: api.build
	./bin/api

###############################################################################
#                                    Tools                                    #
###############################################################################

lint:
	golangci-lint run ./...

test:
	go test ./...
