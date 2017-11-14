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

ENV PATH /usr/local/go/bin:/usr/lib/google-cloud-sdk/bin:/opt/go/bin:${PATH}
RUN echo "PATH=\"${PATH}\"" > /etc/environment

# Add entrypoint that runs bootstrap with appropriate arguments
ADD docker/istio_builders/istio_builder/entrypoint.sh /usr/local/bin/entrypoint
RUN sudo chmod +rx /usr/local/bin/entrypoint

RUN touch /etc/passwd && chmod a+rw /etc/passwd

ENTRYPOINT ["entrypoint"]
