FROM alpine:3.9

RUN apk add bind-tools
RUN  apk update && \
     apk add libc6-compat && \
     apk add ca-certificates && \
     apk add logrotate

WORKDIR /
COPY ./keylist.json /
COPY ./run_highway.sh /
RUN chmod +x /run_highway.sh

CMD ["/bin/sh","/run_highway.sh"]

COPY ./highway /
RUN chmod +x /highway