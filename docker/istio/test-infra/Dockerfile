FROM ubuntu:xenial

# Installing
ADD install/* ./
RUN chmod +x *.sh
RUN ./install.sh && rm -rf install*.sh

# Docker in Docker settings
VOLUME /var/lib/docker
EXPOSE 2375

ENV PATH /usr/local/go/bin:/opt/go/bin:${PATH}

# Add entrypoint to start docker
ADD entrypoint.sh /usr/local/bin/entrypoint
RUN chmod +rx /usr/local/bin/entrypoint

# Set CI variable which can be checked by test scripts to verify
# if running in the continuous integration environment.
ENV CI prow

ENTRYPOINT ["entrypoint"]
