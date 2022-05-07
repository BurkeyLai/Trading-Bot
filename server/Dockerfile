#FROM golang:1.17.5-alpine

#RUN mkdir -p /server

#WORKDIR /server

#COPY . .

#RUN go mod download

#RUN go build -o server

#ENTRYPOINT ["./server"]

FROM golang:1.17.5-alpine AS stage1
ENV RUN_PATH=/server PROJ_PATH=/build
RUN mkdir -p $RUN_PATH
WORKDIR $RUN_PATH
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM stage1 AS stage2
USER root
ADD . $PROJ_PATH
WORKDIR $PROJ_PATH
RUN go build -o server
RUN tar -cvzf server.tar.gz server
RUN tar -zxf server.tar.gz -C $RUN_PATH

FROM alpine
USER root
ENV RUN_PATH=/server
RUN mkdir -p $RUN_PATH
COPY --from=stage2 ${RUN_PATH} ${RUN_PATH}
WORKDIR ${RUN_PATH}
ENTRYPOINT ["./server"]