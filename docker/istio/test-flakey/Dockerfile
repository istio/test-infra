FROM ubuntu:xenial

# Installing
COPY shared/tools /tmp/tools/
COPY test-flakey/install.sh /tmp/tools/
RUN chmod -R +x /tmp/tools/*.sh
WORKDIR /tmp/tools
RUN ./install.sh
WORKDIR /tmp
RUN rm -rf /tmp/tools

# Docker in Docker settings
VOLUME /var/lib/docker
EXPOSE 2375

ENV PATH /usr/local/go/bin:/opt/go/bin:/usr/lib/google-cloud-sdk/bin:${PATH}

# Add entrypoint to start docker
COPY shared/prow-runner.sh /usr/local/bin/entrypoint
RUN chmod +rx /usr/local/bin/entrypoint

# Set CI variable which can be checked by test scripts to verify
# if running in the continuous integration environment.
ENV CI prow

ENTRYPOINT ["entrypoint"]
