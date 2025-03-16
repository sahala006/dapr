FROM centos as builder

WORKDIR /root/dapr

COPY daprd .

copy config.yaml .

CMD ./daprd

EXPOSE 3500

# go build -tags "allcomponents" -o daprd cmd/daprd/main.go
# docker build -t 127.0.0.1:5100/devops/dapr:v4 .