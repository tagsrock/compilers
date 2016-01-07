FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

# Install Dependencies
RUN apt-get update && apt-get install -qy \
  --no-install-recommends \
  ca-certificates \
  # libgmp-dev \
  # libc6-dev \
  && rm -rf /var/lib/apt/lists/*
ENV INSTALL_BASE /usr/local/bin

# Golang
ENV GOLANG_VERSION 1.4.2
RUN curl -sSL https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz \
  | tar -v -C /usr/src -xz
RUN cd /usr/src/go/src && ./make.bash --no-clean 2>&1
ENV PATH /usr/src/go/bin:$PATH
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go

# Go wrapper
ENV GO_WRAPPER_VERSION 1.4
RUN curl -sSL -o $INSTALL_BASE/go-wrapper https://raw.githubusercontent.com/docker-library/golang/master/$GO_WRAPPER_VERSION/wheezy/go-wrapper
RUN chmod +x $INSTALL_BASE/go-wrapper

# Install eris-compilers, a go app that manages compilations
ENV REPO github.com/eris-ltd/eris-compilers
ENV BASE $GOPATH/src/$REPO
ENV NAME eris-compilers
RUN mkdir --parents $BASE
COPY . $BASE/
RUN cd $BASE/cmd/$NAME && go build -o $INSTALL_BASE/$NAME

# Setup User
ENV USER eris
ENV ERIS /home/$USER/.eris

# Add Gandi certs for eris
COPY docker/gandi2.crt /data/gandi2.crt
COPY docker/gandi3.crt /data/gandi3.crt
RUN chown --recursive $USER /data

# Copy in start script
COPY docker/start.sh /home/$USER/

# Point to the compiler location.
RUN mkdir --parents $ERIS/languages
COPY docker/config.json $ERIS/languages/config.json
RUN chown --recursive $USER:$USER /home/$USER

# Finalize
USER $USER
VOLUME $ERIS
WORKDIR /home/$USER
EXPOSE 9098 9099
CMD ["/home/eris/start.sh"]
