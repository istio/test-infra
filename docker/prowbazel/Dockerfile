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
ENV HOME /home/bootstrap
RUN useradd -c "Bootstrap user" -d ${HOME} -G docker,sudo -m bootstrap -s /bin/bash
USER bootstrap
WORKDIR ${HOME}

# Explicitly set bazel cache, to avoid cache issues with checking out
# code and running bazel in different dir than the calling script.
ENV TEST_TMPDIR /home/bootstrap/.cache/bazel

# Move preloaded Bazel binaries to bootstrap's cache.
RUN mkdir -p /home/bootstrap/.cache/ \
    && sudo mv /root/.cache/bazelisk ${HOME}/.cache/ \
    && sudo chown -R bootstrap:bootstrap ${HOME}/.cache/

# Add bootstrap, the test harness (checks out code, captures log and status)
ADD https://raw.githubusercontent.com/nlandolfi/test-infra-1/d7900388fa1fb69738e231a21230373fd4b5ede9/jenkins/bootstrap.py ${HOME}
RUN sudo chmod +rx ${HOME}/bootstrap.py

# Add entrypoint that runs bootstrap with appropriate arguments
ADD ./docker/prowbazel/entrypoint /usr/local/bin/entrypoint
RUN sudo chmod +rx /usr/local/bin/entrypoint

# Final GOPATH configuration
RUN mkdir -p /home/bootstrap/go/src
ENV GOPATH=/home/bootstrap/go
ENV PATH ${GOPATH}/bin:${PATH}

# Set CI variable which can be checked by test scripts to verify
# if running in the continuous integration environment.
ENV CI bootstrap

ENTRYPOINT ["entrypoint"]
