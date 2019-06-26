FROM ubuntu:xenial

# Installing
ADD shared/tools istio/install.sh /tmp/tools/
RUN chmod -R +x /tmp/tools/*.sh
RUN cd /tmp/tools && ./install.sh && rm -rf /tmp/tools

COPY --from=koalaman/shellcheck-alpine:v0.6.0 /bin/shellcheck /bin/shellcheck

# Docker in Docker settings
VOLUME /var/lib/docker
EXPOSE 2375

ENV PATH /usr/local/go/bin:/opt/go/bin:/usr/lib/google-cloud-sdk/bin:/root/.local/bin:${PATH}

# Add entrypoint to start docker
ADD shared/prow-runner.sh /usr/local/bin/entrypoint
RUN chmod +rx /usr/local/bin/entrypoint

# Set CI variable which can be checked by test scripts to verify
# if running in the continuous integration environment.
ENV CI prow

ENTRYPOINT ["entrypoint"]
