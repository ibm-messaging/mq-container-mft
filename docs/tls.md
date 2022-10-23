# IBM MQ Managed File Transfer Agent in Container

# Setup secure connections to queue manager

You can now configure agent and commands in container to connect securely to queue manager(s) using one way TLS security. mTLS is not supported, yet.

Public key of the queue manager(s) must be mounted into agent container and also cipherspec name must be specified through environment variable. The certificates must be supplied in the following directories:

`/etc/mqmft/pki/coordination.crt` - for certificate of coordination queue manager
`/etc/mqmft/pki/command.crt` - for certificate of command queue manager
`/etc/mqmft/pki/agent.crt` - for certificate of agent queue manager

Additionally the cipherspec name for each queue manager must be specified through the following environment variables.
`MFT_COORD_QMGR_CIPHER` - Name of the ciphespec for coordination queue manager
`MFT_CMD_QMGR_CIPHER` - Name of the ciphespec for command queue manager
`MFT_AGENT_QMGR_CIPHER` - Name of the ciphespec for agent queue manager

Agent or commands will not use secure connections if cipherspec names are not specified. For example if `MFT_COORD_QMGR_CIPHER` environment is not specified, then commands that connect to coordination queue manager will not use TLS connections.

Example podman run:

```
podman run \
  -v ./mftlab/agent:/mftagentcfg \
  -v ./srcdir:/mountpath \
  --env MFT_AGENT_NAME=SRCAGENT \
  --env LICENSE=accept \
  --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentconfig.json\
  --env MFT_COORD_QMGR_CIPHER=ECDHE_RSA_AES_256_CBC_SHA384 \
  --env MFT_CMD_QMGR_CIPHER=ECDHE_RSA_AES_256_CBC_SHA384 \
  --env MFT_AGENT_QMGR_CIPHER=ECDHE_RSA_AES_256_CBC_SHA384 \
  --name srcagent \
  icr.io/ibm-messaging/mqmft:latest

```
