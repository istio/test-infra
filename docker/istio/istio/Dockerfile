FROM ubuntu:xenial

# Installing
COPY shared/tools /tmp/tools/
COPY istio/install.sh /tmp/tools/
RUN chmod -R +x /tmp/tools/*.sh
WORKDIR /tmp/tools
RUN ./install.sh
WORKDIR /tmp
RUN rm -rf /tmp/tools

# ShellCheck linter
ENV SHELLCHECK_VERSION=v0.7.0
ADD https://storage.googleapis.com/shellcheck/shellcheck-${SHELLCHECK_VERSION}.linux.x86_64.tar.xz /tmp
RUN tar -xJf /tmp/shellcheck-${SHELLCHECK_VERSION}.linux.x86_64.tar.xz -C /tmp
RUN mv /tmp/shellcheck-${SHELLCHECK_VERSION}/shellcheck bin/shellcheck

# Docker in Docker settings
VOLUME /var/lib/docker
EXPOSE 2375

ENV PATH /usr/local/go/bin:/opt/go/bin:/usr/lib/google-cloud-sdk/bin:/root/.local/bin:${PATH}

# Add entrypoint to start docker
COPY shared/prow-runner.sh /usr/local/bin/entrypoint
RUN chmod +rx /usr/local/bin/entrypoint

# Set CI variable which can be checked by test scripts to verify
# if running in the continuous integration environment.
ENV CI prow

ENTRYPOINT ["entrypoint"]
