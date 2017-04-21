###############################################################################
##  Name:   Dockerfile
##  Date:   2017-02-28
##  Developer:  Chris Page
##  Email:  christopher.page@ca.com
##  Purpose:   This Dockerfile contains the IoT Blinkt Swarm example
################################################################################
FROM alexellis2/go-armhf:1.7.4

# Set the file maintainer (your name - the file's author)
MAINTAINER Page, Christopher T <Christopher.Page@ca.com>

# Install dependencies for Git
RUN apk --update add git

# Add the core
COPY app.go	.

# Pull the Go dependencies
RUN go get -d -v

# Build the package
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

EXPOSE 8080

ENTRYPOINT []

CMD ["./app"]
