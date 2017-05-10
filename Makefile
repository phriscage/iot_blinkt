default: run

run: build
	docker-compose up

build:
	docker-compose -f docker-compose.build.yml build

clean:
	docker-compose stop && docker-compose rm -f

#swarm_run: build

	#docker --host localhost:2375 run --rm -ti -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:${PWD} ddrozdov/docker-compose-swarm-mode -f ${PWD}/docker-compose.yml up

swarm_dry_run:

	docker run --rm -ti -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:${PWD} ddrozdov/docker-compose-swarm-mode -f ${PWD}/docker-compose.yml --dry-run up

#swarm_clean:

	#docker --host localhost:2375 run --rm -ti -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:${PWD} ddrozdov/docker-compose-swarm-mode -f ${PWD}/docker-compose.yml rm -f
