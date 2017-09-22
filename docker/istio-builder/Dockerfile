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

ENV PATH /usr/local/go/bin:/usr/lib/google-cloud-sdk/bin:${PATH}
RUN echo "PATH=\"${PATH}\"" > /etc/environment

# Create bootstrap user
RUN useradd -c "Releaser" -d /home/releng -G docker,sudo -m releng -s /bin/bash
