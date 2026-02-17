# Security

## Container runtime

### User

The MQ MFT agent container image runs with UID 1001(non-root user), though this can be any UID, with a fixed GID of 0 (root).

### Capabilities

The MQ MFT agent container image requires no Linux capabilities, so you can drop any capabilities which are added by default.  For example, in Docker you could do the following:

```sh
docker run \
  --cap-drop=ALL \
  --env LICENSE=accept \
  --env MFT_AGENT_NAME=AGENT1 \
  --env MFT_AGENT_CONFIG_FILE=/config/agent-config.json \
  --detach \
  --name agentone \
  icr.io/ibm-messaging/mqmft:latest
```
