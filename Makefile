SHELL := /bin/bash

default: run

run: build
	docker-compose up

build:
	docker-compose -f docker-compose.build.yml build

clean:
	docker-compose stop && docker-compose rm -f
