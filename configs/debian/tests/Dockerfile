FROM debian:bookworm

RUN apt-get update \
    && apt-get install -y bats bats-assert bats-file \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY . /tmp/

CMD [ "/tmp/test.bats" ]
