FROM golang:1.13.1

LABEL maintainer="Hoang Nguyen Gia <hoang@incognito.org>"

RUN apt-get -y update
RUN apt-get -y install supervisor
RUN apt-get -y install net-tools

COPY docker/supervisor.d/* /etc/supervisor/conf.d/worker/
COPY docker/supervisord.conf /etc/supervisor/supervisord.conf

WORKDIR /go/src/app
COPY go.mod .
RUN go mod download

COPY . .

RUN go build -o highway

EXPOSE 9330
CMD ["/usr/bin/supervisord"]