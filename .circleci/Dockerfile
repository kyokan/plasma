FROM ubuntu:bionic

RUN apt-get update && \
    apt-get install -y curl gnupg && \
    curl -sSL https://deb.nodesource.com/gpgkey/nodesource.gpg.key | apt-key add && \
    echo "deb https://deb.nodesource.com/node_11.x bionic main" | tee /etc/apt/sources.list.d/nodesource.list && \
    echo "deb-src https://deb.nodesource.com/node_11.x bionic main" | tee -a /etc/apt/sources.list.d/nodesource.list && \
    apt-get update && \
    apt-get install -y build-essential git ca-certificates gzip nodejs

RUN curl -o go.tgz https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz && \
    echo "ff54aafedff961eb94792487e827515da683d61a5f9482f668008832631e5d25 go.tgz" | sha256sum -c && \
    tar -C /usr/local -xzvf go.tgz && \
    rm go.tgz && \
    export PATH="/usr/local/go/bin:$PATH" && \
    go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH" && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh -o install.sh && \
    echo "9507d8826574a5b25cf069ab9311793e5d5fc88bba3bdfd02131fae8f50ed1bc  install.sh" | sha256sum -c && \
    sh install.sh && \
    echo "export PATH=$PATH" > /etc/environment

WORKDIR $GOPATH