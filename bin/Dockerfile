FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install -y ca-certificates cronolog cron
RUN apt-get install -y dnsutils
WORKDIR /
COPY removeoldlog /etc/cron.d/removeoldlog
RUN chmod 0644 /etc/cron.d/removeoldlog
RUN crontab /etc/cron.d/removeoldlog

COPY ./keylist.json /
COPY ./run_highway.sh /
RUN chmod +x /run_highway.sh

CMD ["/bin/sh","/run_highway.sh"]

COPY ./highway /
RUN chmod +x /highway