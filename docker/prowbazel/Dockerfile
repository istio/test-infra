FROM ubuntu:xenial

# Note: The build directory is the root of the istio/test-infra repo, not ./

# Installing necessary packages
RUN rm -rf /var/lib/apt/lists/* \
    && apt-get update --fix-missing -qq \
    && apt-get install -qqy git iptables procps sudo xz-utils \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Adding sudo group user no password access.
# This is used by bootstrap user to start docker service
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# Installing
ADD scripts /tmp/istio_tmp/scripts
RUN chmod +x /tmp/istio_tmp/scripts/linux-install-software
RUN /tmp/istio_tmp/scripts/linux-install-software \
    && rm -rf /tmp/istio_tmp

VOLUME /var/lib/docker
EXPOSE 2375

ENV PATH /usr/local/go/bin:/usr/local/cmake/bin:/usr/lib/google-cloud-sdk/bin:/opt/go/bin:${PATH}

# Create bootstrap user
ENV HOME /
RUN useradd -c "Bootstrap user" -d ${HOME} -G docker,sudo -m bootstrap -s /bin/bash
USER bootstrap
WORKDIR ${HOME}

# Explicitly set bazel cache, to avoid cache issues with checking out
# code and running bazel in different dir than the calling script.
ENV TEST_TMPDIR /.cache/bazel

# Add entrypoint that runs bootstrap with appropriate arguments
# Final GOPATH configuration
ENV GOPATH=/go
ENV PATH ${GOPATH}/bin:${PATH}

# Set CI variable which can be checked by test scripts to verify
# if running in the continuous integration environment.
ENV CI bootstrap

