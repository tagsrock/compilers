FROM ubuntu:14.04
MAINTAINER Eris Industries <support@erisindustries.com>

# Install Golang
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get install -qy \
  ca-certificates \
  curl \
  gcc \
  git \
  libgmp-dev \
  software-properties-common \
  libc6-dev
ENV GOLANG_VERSION 1.4.2
RUN curl -sSL https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz \
  | tar -v -C /usr/src -xz
RUN cd /usr/src/go/src && ./make.bash --no-clean 2>&1
ENV PATH /usr/src/go/bin:$PATH
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go

# Compilers
RUN add-apt-repository -y ppa:ethereum/ethereum && \
  add-apt-repository -y ppa:ethereum/ethereum-dev && \
  apt-get update && apt-get install -qy \
  lllc \
  sc \
  solc \
  && rm -rf /var/lib/apt/lists/*

# LLLC-server, a go app that manages compilations
ENV repository lllc-server
RUN mkdir --parents $GOPATH/src/github.com/eris-ltd/$repository
COPY . $GOPATH/src/github.com/eris-ltd/$repository
WORKDIR $GOPATH/src/github.com/eris-ltd/$repository/cmd/$repository
RUN go install

# Add Gandi certs for eris
COPY docker/gandi2.crt /data/gandi2.crt
COPY docker/gandi3.crt /data/gandi3.crt

# Add Eris User
RUN groupadd --system eris && useradd --system --create-home --gid eris eris

# Copy in start script
COPY docker/start.sh /home/eris/

# Point to the compiler location.
RUN mkdir --parents /home/eris/.eris/languages
COPY docker/config.json /home/eris/.eris/languages/config.json
RUN chown --recursive eris /home/eris/.eris
RUN chown --recursive eris /data

USER eris
WORKDIR /home/eris/

EXPOSE 9098 9099
CMD ["/home/eris/start.sh"]
