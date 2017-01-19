# Stored at gcr.io/endpoints-jenkins/ubuntu-16-04-slave:latest
FROM ubuntu:xenial

# Installing bazel deps
RUN rm -rf /var/lib/apt/lists/* \
    && apt-get update --fix-missing -qq \
    && apt-get install -qqy \
        g++ \
	gcc \
	git pkg-config \
	openjdk-8-jdk \
	unzip \
	wget \
	zip \
	zlib1g-dev \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Installing Tools
ADD scripts /tmp/esp_tmp/scripts
ADD docker/bazel-remote/linux-install-bazel-remote \
    /tmp/esp_tmp/scripts/
ADD docker/bazel-remote/entrypoint /usr/local/bin/entrypoint

RUN chmod +rx /usr/local/bin/entrypoint \
    /tmp/esp_tmp/scripts/linux-install-bazel-remote

RUN /tmp/esp_tmp/scripts/linux-install-bazel-remote \
    && rm -rf /tmp/esp_tmp

ENV PATH /usr/local/bin:${PATH}

ENTRYPOINT ["entrypoint"]
