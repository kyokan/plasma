FROM golang:1.11.1-alpine3.8

RUN apk add git make nodejs npm python g++ curl && \
    mkdir -p $GOPATH/src/github.com/kyokan/plasma && \
    cd $GOPATH/src/github.com/kyokan/plasma && \
    git clone https://github.com/kyokan/plasma . && \
    git checkout build && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    npm i -g truffle && \
    go get -u github.com/ethereum/go-ethereum && \
    cd $GOPATH/src/github.com/ethereum/go-ethereum && \
    go install ./cmd/abigen

RUN cd $GOPATH/src/github.com/kyokan/plasma && \
    make deps && \
    GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o ./target/plasma ./cmd/plasma/main.go

CMD ["echo", "Done."]