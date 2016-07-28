FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

# Install Dependencies
RUN apt-get update && apt-get install -qy \
  --no-install-recommends \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

ENV INSTALL_BASE /usr/local/bin

# GOLANG
ENV GOLANG_VERSION 1.6
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 5470eac05d273c74ff8bac7bef5bad0b5abbd1c4052efbdbc8db45332e836b0b
RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR /go

# GO WRAPPER
ENV GO_WRAPPER_VERSION 1.6
RUN curl -sSL -o $INSTALL_BASE/go-wrapper https://raw.githubusercontent.com/docker-library/golang/master/$GO_WRAPPER_VERSION/wheezy/go-wrapper
RUN chmod +x $INSTALL_BASE/go-wrapper

# GLIDE INSTALL
RUN sudo add-apt-repository ppa:masterminds/glide \
  && sudo apt-get update
RUN sudo apt-get install glide

# Install eris-compilers, a go app that manages compilations
ENV REPO github.com/eris-ltd/eris-compilers
ENV BASE $GOPATH/src/$REPO
ENV NAME eris-compilers
RUN mkdir --parents $BASE
COPY . $BASE/
RUN cd $BASE && glide install && \
  cd $BASE/cmd/$NAME && go install ./
RUN unset GOLANG_VERSION && \
  unset GOLANG_DOWNLOAD_URL && \
  unset GOLANG_DOWNLOAD_SHA256 && \
  unset GO_WRAPPER_VERSION && \
  unset REPO && \
  unset BASE && \
  unset NAME && \
  unset INSTALL_BASE

# Setup User
ENV USER eris
ENV ERISDIR /home/$USER/.eris

# Add Gandi certs for eris
COPY docker/gandi2.crt /data/gandi2.crt
COPY docker/gandi3.crt /data/gandi3.crt
RUN chown --recursive $USER /data

# Copy in start script
COPY docker/start.sh /home/$USER/

# Point to the compiler location.
RUN chown --recursive $USER:$USER /home/$USER

# Finalize
USER $USER
VOLUME $ERISDIR
WORKDIR /home/$USER
EXPOSE 9098 9099
CMD ["/home/eris/start.sh"]
