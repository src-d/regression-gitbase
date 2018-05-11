FROM golang:1.10-stretch
MAINTAINER source{d}

ENV LOG_LEVEL=debug
ENV REG_REPOS=/cache/repos
ENV REG_BINARIES=/cache/binaries

RUN apt-get update && \
    apt-get install -y dumb-init \
      git make bash gcc libxml2-dev && \
    apt-get autoremove -y && \
    ln -s /usr/local/go/bin/go /usr/bin

ADD build/regression-gitbase_linux_amd64/regression /bin/

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/regression", "latest", "remote:master"]
