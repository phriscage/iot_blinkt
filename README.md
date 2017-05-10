# IoT - Blinkt

An IoT Docker Swarm environment with Raspberry Pi Zero IoT devices and [Blinkt!](https://shop.pimoroni.com/products/blinkt)

## Prerequisites

* Docker environment running
* [Docker Swarm knowledge](https://www.docker.com/products/docker-swarm)
* [Rasperry Pi Rasbian OS and OTG Configuration](https://cawiki.ca.com/display/~pagch04/Raspberry+Pi+Rasbian+OS+Configuration)
* Download this [repo](https://github.com/phriscage/iot_blinkt)

### Setup
There are three main components to setup the environment.
* 	Configure all IoT worker nodes with unique hostnames
*	Configure Docker Swarm and create worker nodes
* 	Create the sample service from the docker-compose configuration and verify the status
*	Consume and scale the service


#### Configure IoT Worker nodes
The IoT Raspberry Pi worker nodes are privisioned with the latest Rasbian OS, OTG networking, SSH, and I2C GPIO are enabled. Static MACs are recommended for interfaces and gateway to avoid duplicate RDNIS MAC OSX interface renaming. Plug one Raspberry Pi USB into your MACOSX and create a custom hostname to avoid duplicates. Bonjour should pick up the new hostname and then you can add subsequent Raspberry Pi workers. Take note of the hostnames so you can iterate over them during the Docker Swarm joining.


#### Configure Docker Swarm for Rasbian IoT nodes
The Docker Swarm first requires swarm initialization with a master node. For this tutorial, we will use the local Docker machine on OSX for the master node. Multi-node swarm setup is currently not-available for native Docker for MAC/Windows, https://docs.docker.com/engine/swarm/swarm-tutorial/, so we will use a [DIND](https://hub.docker.com/_/docker/) instance as the master node. For DIND, separate Docker-dind containers are created to simulate a separate Docker machine. So you can essentially run the Docker commands in a separate Docker environment inside your parent Docker environment. Subsequent Raspberry PI worker nodes (Raspberry Pi 3/Zero) are next bootstrapped and joined to the cluster. A Raspberry PI could play the role of master instead of OSX.

Create the DIND master node container:

We need to expose the Swarm ports and change the advertise address so worker nodes can communicate with the DIND swarm master. TCP 2375 for docker API communication, TCP 2377 for Docker swarm. Any additional ports for containers/applications will need to be exposed to the DIND container when it is created. Do not mount the '/var/lib/docker' volume locally at this time. The current DIND image is *docker:17.03-dind*:

	docker run -d --privileged --name manager --hostname=manager -p 2375:2375 -p 2377:2377 -p 7946:7946 -p 4789:4789 -p 8000:8000 -p 8080:8080 docker:17.03-dind;

Initialize the Docker Swarm environment

Since Docker for MAC/Windows runs natively, typically you connect via localhost. We need to utilize a reachable IP address from the worker nodes instead of 127.0.0.1. My IoT devices are on an OTG bridge via USB so en8 and bridge100 when sharing access to the internet. `ifconfig bridge100 | grep "inet " | awk '{print $2}'`

	docker -H localhost:2375 swarm init --advertise-addr `ifconfig bridge100 | grep "inet " | awk '{print $2}'`

Create a Swarm token environment variable

    SWARM_TOKEN=$(docker -H localhost:2375 swarm join-token -q worker)

Create a Swarm master IP address environment variable

    SWARM_MASTER=$(docker -H localhost:2375 info | grep -w 'Node Address' | awk '{print $3}')

Loop through the IoT worker hostnames via the IOT_WORKERS environment variable and execute the swarm join command with SWARM_TOKEN and SWARM_MASTER variables defined above.

	IOT_WORKERS="thing2.local thing3.local";
	IOT_USERNAME="pi";
	for host in $IOT_WORKERS; do
		ssh $IOT_USERNAME@$host "docker swarm leave; docker swarm join --token $SWARM_TOKEN $SWARM_MASTER:2377"
	done

Verify the nodes are now part of the Swarm:

	docker -H localhost:2375 node ls


#### Configure Swarm visualizer
Swarm visualizer is nice Express JS app to view the Swarm environment. Start the Swarm visualization service in a separate terminal point your browser to http://localhost:8000. You will need to download the image to the DIND environment or mount a volume instead.

	docker -H localhost:2375 service create --with-registry-auth  --name visualizer --constraint 'node.role==manager'  --mount 'type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock,readonly=0'  --publish '8000:8080' manomarks/visualizer

Point to http://localhost:8000 to view the nodes and services vi visualizer


#### Launch the IoT application with Blinkt service
Start the Blinkt service with Docker stack deploy. Point to the docker-compose.yml and add a service name prefix: 'iot'

	docker -H localhost:2375 stack deploy -c docker-compose.yml iot

Check the service status

	docker -H localhost:2375 service ls

Verify the service is accessible from the manager and from the worker:

	curl -4 -i http://localhost:8080
	curl -4 -i http://thing2.local:8080

Run some curl commands to illuminate the LEDs:

	while ((1)); do for color in red blue green; do curl -4 -i -X POST "http://thing2.local:8080/blinkts/random?delay=10&color=${color}"; sleep 0.1; done; done

Scale the service

	docker -H localhost:2375 service scale iot_blinkt=2

Point to http://localhost:8080 to view the nodes and services
