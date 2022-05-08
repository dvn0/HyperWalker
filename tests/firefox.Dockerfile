FROM debian:bullseye-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -yq && \
    apt-get install -yq \
            --no-install-recommends \
              firefox-esr
