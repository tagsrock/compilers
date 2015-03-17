# Ethereum C++ Client
# https://github.com/ethereum/cpp-ethereum/blob/74b085efea90dbce18a7320ff05cda0de0bd99d5/docker/Dockerfile

# Make sure you have at least 2 GB RAM.

FROM ubuntu:14.04

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update
RUN apt-get upgrade -y

# Ethereum dependencies
RUN apt-get install -qy build-essential g++-4.8 git cmake libboost-all-dev libcurl4-openssl-dev wget
RUN apt-get install -qy automake unzip libgmp-dev libtool libleveldb-dev yasm libminiupnpc-dev libreadline-dev scons
RUN apt-get install -qy libjsoncpp-dev libargtable2-dev

# NCurses based GUI (not optional though for a succesful compilation, see https://github.com/ethereum/cpp-ethereum/issues/452 )
RUN apt-get install -qy libncurses5-dev

# Qt-based GUI
# RUN apt-get install -qy qtbase5-dev qt5-default qtdeclarative5-dev libqt5webkit5-dev

# Ethereum PPA
RUN apt-get install -qy software-properties-common
RUN add-apt-repository ppa:ethereum/ethereum
RUN apt-get update
RUN apt-get install -qy libcryptopp-dev libjson-rpc-cpp-dev

# Go
# https://github.com/docker-library/golang/blob/304244cfe374d3477e746c057a5ec4dcf7a47657/1.4/Dockerfile

RUN apt-get install -qy curl

# gcc for cgo
RUN apt-get update && apt-get install -y \
    gcc libc6-dev make \
    --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.4.1

RUN curl -sSL https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz \
    | tar -v -C /usr/src -xz

RUN cd /usr/src/go/src && ./make.bash --no-clean 2>&1

ENV PATH /usr/src/go/bin:$PATH

RUN mkdir -p /go/src
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go

# Install LLL (eris LLL, which includes a few extra opcodes)

ENV repository eris-cpp
RUN mkdir /usr/local/src/$repository
WORKDIR /usr/local/src/$repository

RUN curl --location https://github.com/eris-ltd/$repository/archive/ac00bf784a1c5094e65887e41f03fb735223d9df.tar.gz \
 | tar --extract --gzip --strip-components=1

WORKDIR build
RUN bash instructions

# LLLC-server

ENV repository lllc-server
RUN mkdir --parents $GOPATH/src/github.com/eris-ltd/$repository
# WORKDIR $GOPATH/src/github.com/eris-ltd/$repository

# RUN curl --location https://github.com/eris-ltd/$repository/archive/d482eddefba8db221b7d5031698a54c2127a138d.tar.gz \
 # | tar --extract --gzip --strip-components=1
# RUN go get .

COPY . $GOPATH/src/github.com/eris-ltd/$repository
WORKDIR $GOPATH/src/github.com/eris-ltd/$repository/cmd/lllc-server
RUN go get -d ./... && go install

ENV user eris
RUN groupadd --system $user && useradd --system --create-home --gid $user $user

## Point to the compiler location.

COPY tests/config.json /home/$user/.decerver/languages/
RUN chown --recursive $user /home/$user/.decerver

# Test compiler.
USER $user
WORKDIR ../..
RUN go test --test.run LLLClientLocal

EXPOSE 5000
CMD ["lllc-server"]
