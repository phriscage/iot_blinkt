# IoT - Blinkt

An IoT Docker Swarm environment with Raspberry Pi Zero IoT devices and [Blinkt!](https://shop.pimoroni.com/products/blinkt)


## Prerequisites

* Docker environment running
* [Docker Swarm knowledge](https://www.docker.com/products/docker-swarm)
* [Rasperry Pi Rasbian OS Configuration](https://cawiki.ca.com/display/~pagch04/Raspberry+Pi+Rasbian+OS+Configuration)
* Download this [repo](https://github.com/phriscage/iot_blinkt)


### Setup
There are three main components to setup the environment.
* 	Configure all IoT worker nodes with unique hostnames
*	Configure Docker Swarm and create worker nodes
* 	Create the sample service from the docker-compose configuration and verify the status
*	Consume and scale the service


#### Configure Docker Swarm for Rasbian
The Docker Swarm first requires swarm initialization with a master node. For this tutorial, we will use the local Docker machine on OSX for the master node. Multi-node swarm setup is currently not-available for native Docker for MAC/Windows, https://docs.docker.com/engine/swarm/swarm-tutorial/, so we will use a [DIND](https://hub.docker.com/_/docker/) instance as the master node. Subsequent IoT worker nodes are next created and joined to the cluster. IoT middle nodes could also be a master and will be added at a later date.

Create the DIND container:

We need to expose the Swarm ports and change the advertise address so worker nodes can communicate with the DIND swarm master. TCP 2375 for docker API communication, TCP 2377 for Docker swarm. Any additional ports for containers/applications will need to be exposed to the DIND container when it is created. Do not mount the '/var/lib/docker' volume locally at this time.

	docker run -d --privileged --name manager --hostname=manager -p 2375:2375 -p 2377:2377 -p 80:80 -p 443:443 -p 8000:8000 -p 8080:8080 docker:1.13-dind;

You van view the DIND containers that are running on the parent Docker machine

	docker ps -a

Determine the correct IP address to advertise:

Since Docker for MAC/Windows runs natively, typicall you connect via localhost. We need to utilize a reachable IP address from the worker nodes instead of 127.0.0.1. My IoT devices are on an OTG bridge via USB so en8 and bridge100 when sharing access to the internet

	ipconfig getifaddr en8
	ifconfig bridge100 | grep "inet " | awk '{print $2}'

Initialize the Docker Swarm environment

	docker --host localhost:2375 swarm init --advertise-addr `ifconfig bridge100 | grep "inet " | awk '{print $2}'`

Capture the Swarm join command with token

	docker --host localhost:2375 swarm join-token -q worker

On each IoT worker shell session, exectute the swarm join command above (Change the token and swarm master ips).

	ssh pi@thing1.local; docker swarm join --token SWMTKN-1-4bj6bn7bral406aqofau8ikry0yna4wt0e5c542f0etbcgvx9a-7z47bebkn9cl7lpbrchoemv2d 192.168.2.1:2377

Verify the nodes are now part of the Swarm:

	docker --host localhost:2375 node ls


#### Launch the Blinkt service


#### Configure Swarm visualizer
Swarm visualizer is nice Express JS app to view the Swarm environment. Start the Swarm visualization service in a separate terminal point your browser to http://localhost:8000. You will need to download the image to the DIND environment or mount a volume instead.

	docker --host localhost:2375 service create --with-registry-auth  --name visualizer --constraint 'node.role==manager'  --mount 'type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock,readonly=0'  --publish '8080:8080' manomarks/visualizer

Check the service status

	docker --host localhost:2375 service ls

Point to http://localhost:8080 to view the nodes and services
