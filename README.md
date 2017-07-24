# IoT - Blinkt

An IoT Docker Swarm environment with Raspberry Pi Zero IoT devices and [Blinkt!](https://shop.pimoroni.com/products/blinkt)

## <a name="prerequisites"></a>Prerequisites:

*	[Rasperry Pi](https://www.raspberrypi.org)
*	[Rasperry Pi Rasbian OS installed](https://www.raspberrypi.org/documentation/installation/installing-images/)
*	[Rasperry Pi Rasbian OTG Configured](https://gist.github.com/gbaman/975e2db164b3ca2b51ae11e45e8fd40a)
*	[Docker Swarm knowledge](https://www.docker.com/products/docker-swarm)
*	Docker environment running
*	Download this [repo](https://github.com/phriscage/iot_blinkt)


### <a name="configuration"></a>Configuration:

* 	[Nodes](#configure_nodes) - Configure all IoT worker nodes with unique hostnames
*	[Swarm](#configure_swarm) - Configure Docker Swarm and create worker nodes
*	[Visualizer](#configure_visualizer) - Configure Docker Swarm [Visualizer](https://github.com/dockersamples/docker-swarm-visualizer) to view the nodes


### <a name="installation"></a>Installtion:

* 	[Deploy](#deploy) - Create the sample Blinkt service from the docker-compose configuration and verify the status
*	[Consume](#consume) - Consume and scale the service


#### <a name="configure_nodes"></a>Configure IoT Worker nodes:

The IoT Raspberry Pi worker nodes are privisioned with the latest Rasbian OS, OTG networking, SSH, and I2C GPIO are enabled. Static MACs are recommended for interfaces and gateway to avoid duplicate RDNIS MAC OSX interface renaming. Plug one Raspberry Pi USB into your MACOSX and create a custom hostname to avoid duplicates. Bonjour should pick up the new hostname and then you can add subsequent Raspberry Pi workers. Take note of the hostnames so you can iterate over them during the Docker Swarm joining.


#### <a name="configure_swarm"></a>Configure Docker Swarm for Rasbian IoT nodes:

The Docker Swarm first requires swarm initialization with a master node. For this tutorial, we will use the local Docker machine on OSX for the master node. Multi-node swarm setup is currently not-available for native Docker for MAC/Windows, https://docs.docker.com/engine/swarm/swarm-tutorial/, so we will use a [DIND](https://hub.docker.com/_/docker/) instance as the master node. For DIND, separate Docker-dind containers are created to simulate a separate Docker machine. So you can essentially run the Docker commands in a separate Docker environment inside your parent Docker environment. Subsequent Raspberry PI worker nodes (Raspberry Pi 3/Zero) are next bootstrapped and joined to the cluster. A Raspberry PI could play the role of master instead of OSX.

Create the DIND master node container:

We need to expose the Swarm ports and change the advertise address so worker nodes can communicate with the DIND swarm master. TCP 2375 for docker API, TCP 2377 and 7946 for Docker Swarm communication. Any additional ports for containers/applications will need to be exposed to the DIND container when it is created. Do not mount the '/var/lib/docker' volume locally at this time. The current DIND image is *docker:17.06-dind*:

	docker run -d --privileged --name manager --hostname=manager -p 2375:2375 -p 2377:2377 -p 7946:7946 -p 7946:7946/udp -p 4789:4789/udp -p 8000:8000 -p 8080:8080 docker:17.06-dind;

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


#### <a name="configure_visualizer"></a>Configure Swarm visualizer:

Swarm visualizer is nice Express JS app to view the Swarm environment. Start the Swarm visualization service in a separate terminal point your browser to http://localhost:8000. You will need to download the image to the DIND environment or mount a volume instead.

	docker -H localhost:2375 service create --with-registry-auth  --name visualizer --constraint 'node.role==manager'  --mount 'type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock,readonly=0'  --publish '8000:8080' manomarks/visualizer

Point to http://localhost:8000 to view the nodes and services vi visualizer


#### <a name="deploy"></a>Launch the IoT application with Blinkt service:

Start the Blinkt service with Docker stack deploy. Point to the docker-compose.yml and add a service name prefix: 'iot'. *--resolve-image=never* is used since the image architecture **arm** is not identified in swarmkit.

	docker -H localhost:2375 stack deploy -c docker-compose.yml --resolve-image=never iot

Check the stack status

	docker -H localhost:2375 stack ps iot

Check the service status

	docker -H localhost:2375 service ls

Verify the service is accessible from the workers. *DIND manager node does not have bridge networking enabled so can't access the service without LB/GW*:

	_curl -4 -i http://localhost:8080_
	curl -4 -i http://thing2.local:8080
	curl -4 -i http://thing3.local:8080

Run some curl commands to illuminate the LEDs:

	while ((1)); do for color in red blue green; do curl -4 -i -X POST "http://thing2.local:8080/blinkts/random?delay=10&color=${color}"; sleep 0.1; done; done

#### <a name="scale"></a>Scale and consume the service:

	docker -H localhost:2375 service scale iot_blinkt=2

Point to http://localhost:8080 to view the nodes and services
