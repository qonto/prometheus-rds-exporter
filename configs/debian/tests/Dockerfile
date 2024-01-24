#checkov:skip=CKV2_DOCKER_1:Sudo is required to test installation and suppression of the exporter Debian package

FROM debian:bookworm

HEALTHCHECK NONE

RUN apt-get update \
    && apt-get install -y bats bats-assert bats-file sudo \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY sudoers /etc/sudoers.d/

RUN useradd -m unittest --groups sudo

USER unittest

COPY . /tmp/

CMD [ "/tmp/test.bats" ]
