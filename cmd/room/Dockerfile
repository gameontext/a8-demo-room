# Derive from Amalgam8's alpine-based sidecar image
FROM amalgam8/a8-sidecar:0.4-alpine

# Amalgam8 sidecar configuration
COPY amalgam8.yaml /opt/chatter/amalgam8.yaml
ENV A8_CONFIG /opt/chatter/amalgam8.yaml

COPY bin/room /opt/chatter/room
EXPOSE 80

# Workaround A8 bug where "--supervise=false" is set on base image
CMD [ "--supervise=true" ]