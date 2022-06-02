FROM golang:bullseye as builder
RUN apt update && apt install -y git make bash jq libhwloc-dev ocl-icd-opencl-dev
RUN  git clone https://github.com/filecoin-project/lotus $GOPATH/src/github.com/filecoin-project/lotus
RUN cd $GOPATH/src/github.com/filecoin-project/lotus && make deps
COPY . app
run mkdir -p $GOPATH/src/github.com/coryschwartz
RUN mv app $GOPATH/src/github.com/coryschwartz/lotus-api-bench
RUN cd $GOPATH/src/github.com/coryschwartz/lotus-api-bench && go build -o /lotus-api-bench main.go

FROM debian:bullseye
RUN apt update && apt install -y libhwloc15 ocl-icd-libopencl1
COPY --from=builder /lotus-api-bench /lotus-api-bench
ENTRYPOINT /lotus-api-bench
