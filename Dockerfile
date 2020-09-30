
# Base our image on an official, minimal image of our preferred golang
FROM golang:1.9

# Note: The default golang docker image, already has the GOPATH env variable set.
# GOPATH is located at /go
ENV GO_SRC $GOPATH/src
ENV kudzu_GITHUB github.com/kudzu-cms/kudzu
ENV kudzu_ROOT $GO_SRC/$kudzu_GITHUB

# Consider updating package in the future. For instance ca-certificates etc.
# RUN apt-get update -qq && apt-get install -y build-essential

# Make the kudzu root directory
RUN mkdir -p $kudzu_ROOT

# All commands will be run inside of kudzu root
WORKDIR $kudzu_ROOT

# Copy the kudzu source into kudzu root.
COPY . .

# the following runs the code inside of the $GO_SRC/$kudzu_GITHUB directory
RUN go get -u $kudzu_GITHUB...

# Define the scripts we want run once the container boots
# CMD [ "" ]
