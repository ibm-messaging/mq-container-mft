
# Running the image with Docker Run
 This document describes the steps for using IBM MQ Managed File Transfer Docker container.

1) Download the image from DockerHub repository
`docker pull docker.io/ibmcom/mqmft:latest`

2) Create a json file, like agentcfg.json containig the information required for creating an agent.  The attributes of the JSON file are described [here](https://github.com/ibm-messaging/mft-cloud/blob/mftubi/docs/agentconfig.md). 
Sample agent configuration JSON is here:
```
{
   "coordinationQMgr":{
      "name":"MFTCORDQM",
      "host":"coordqm.ibm.com",
      "port":1414,
      "channel":"MFTCORDSVRCONN",
      "additionalProperties" : {
         "coordinationQMgrAuthenticationCredentialsFile":"/mftagentcfg/agentcfg/MQMFTCredentials.xml"
      }
   },
   "commandQMgr":{
      "name":"MFTCMDQM",
      "host":"cmdqm.ibm.com",
      "port":1414,
      "channel":"MFTCMDSVRCONN",
      "additionalProperties" : {
         "connectionQMgrAuthenticationCredentialsFile":"/mftagentcfg/agentcfg/MQMFTCredentials.xml"
      }
   },
   "agents":[{
      "name":"SRCAGENT",
      "type":"STANDARD",
      "qmgrName":"MFTAGENTQM",
      "qmgrHost":"agentqm.ibm.com",
      "qmgrPort":1414,
      "qmgrChannel":"MFTSVRCONN",
      "additionalProperties":{
         "enableQueueInputOutput":"true",
         "agentQMgrAuthenticationCredentialsFile":"/mftagentcfg/agentcfg/MQMFTCredentials.xml"
      }
   },
   {
      "name":"AGENTDEST",
      "type":"BRIDGE",
      "qmgrName":"MFTAGENTQM",
      "qmgrHost":"agentqmdest.mycomp.com",
      "qmgrPort":1818,
      "qmgrChannel":"MFTSVRCONN",
      "protocolBridge" : {
         "serverType":"FTP",
         "serverHost":"ftp.ibm.com",
         "serverTimezone":"Europe/London",
         "serverPlatform":"UNIX",
         "serverLocale":"en-GB",
         "serverListFormat"="UNIX", 
         "serverLimitedWrite"="false", 
         "serverFileEncoding"="UTF8", 
         "serverPassiveMode"="true", 
	  },
	  "additionalProperties": {
         "agentQMgrAuthenticationCredentialsFile" : "/mftagentcfg/agentcfg/MQMFTCredentials.xml",
		 "protocolBridgeCredentialConfiguration" : "/mqmftbridgecred/agentcreds/ProtocolBridgeCredentials.prop"
      }
   }]
}
```

3) Create docker volume on host system. This volume needs to be mounted on a container when running.
`docker volume create mftagentcfg`   

4) Determine the path of the volume on local file system
`docker volume inspect mftagentcfg`
   The output would be something like
```
   [
    {
        "CreatedAt": "2021-03-11T03:11:18-07:00",
        "Driver": "local",
        "Labels": {},
        "Mountpoint": "/var/lib/docker/volumes/mfagentcfg/_data",
        "Name": "mftdata",
        "Options": {},
        "Scope": "local"
    }
   ]
```

5) Create a sub-directory under the directory shown in Mountpoint attribute above.
`mkdir /var/lib/docker/volumes/mfagentcfg/_data/agentcfg`
`cp agentcfg.json /var/lib/docker/volumes/mfagentcfg/_data/agentcfg`

6) Create a MQMFTCredentials.xml file and put user mapping for the queue manager you are connecting to.
	For example: Note that "user" attribute has not specified in the entry. This allows agent to pick the credentials for any user that it is running under.
```
<?xml version="1.0" encoding="UTF-8"?>
<tns:mqmftCredentials xmlns:tns="http://wmqfte.ibm.com/MQMFTCredentials" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://wmqfte.ibm.com/MQMFTCredentials MQMFTCredentials.xsd">
    <tns:qmgr name="MFTQM" mqUserId="mftqmuserid" mqPassword="mftqmpassw0rd"/>
</tns:mqmftCredentials>
```
7) Run chmod on MQMFTCredentials.xml file:
`cp MQMFTCredentials.xml /var/lib/docker/volumes/mfagentcfg/_data/agentcfg`

#### Configuration of coordination queue manager
Define following objects in a queue manager.
```
DEFINE TOPIC('SYSTEM.FTE') TOPICSTR('SYSTEM.FTE') REPLACE
ALTER TOPIC('SYSTEM.FTE') NPMSGDLV(ALLAVAIL) PMSGDLV(ALLAVAIL)
DEFINE QLOCAL(SYSTEM.FTE) LIKE(SYSTEM.BROKER.DEFAULT.STREAM) REPLACE
ALTER QLOCAL(SYSTEM.FTE) DESCR('Stream for MQMFT Pub/Sub interface')
DISPLAY NAMELIST(SYSTEM.QPUBSUB.QUEUE.NAMELIST)
ALTER NAMELIST(SYSTEM.QPUBSUB.QUEUE.NAMELIST) +
NAMES(SYSTEM.BROKER.DEFAULT.STREAM+
 ,SYSTEM.BROKER.ADMIN.STREAM,SYSTEM.FTE)
DISPLAY QMGR PSMODE
ALTER QMGR PSMODE(ENABLED)
```

#### Configuration of agent queue manager
Define the following objects in a queue manager. Replace `<AGENTNAME>` with your agent name.

```
DEFINE QLOCAL(SYSTEM.FTE.COMMAND.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(5000) +
    MAXMSGL(4194304) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.DATA.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(5000) +
    MAXMSGL(4194304) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.REPLY.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(5000) +
    MAXMSGL(4194304) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.STATE.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(5000) +
    MAXMSGL(4194304) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.EVENT.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(5000) +
    MAXMSGL(4194304) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.AUTHAGT1.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(0) +
    MAXMSGL(0) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.AUTHTRN1.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(0) +
    MAXMSGL(0) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.AUTHOPS1.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(0) +
    MAXMSGL(0) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.AUTHSCH1.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(0) +
    MAXMSGL(0) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.AUTHMON1.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(0) +
    MAXMSGL(0) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.AUTHADM1.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(0) +
    MAXMSGL(0) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
DEFINE QLOCAL(SYSTEM.FTE.HA.<AGENTNAME>) +
    DEFPRTY(0) +
    DEFSOPT(SHARED) +
    GET(ENABLED) +
    MAXDEPTH(0) +
    MAXMSGL(0) +
    MSGDLVSQ(PRIORITY) +
    PUT(ENABLED) +
    RETINTVL(999999999) +
    SHARE +
    NOTRIGGER +
    USAGE(NORMAL) +
    REPLACE
```	

8) Run the container using docker run command.
  Environment variables to be passed the docker run command
- **LICENSE** - Required. Set this to `accept` to agree to the MQ Advanced for Developers license. If you wish to see the license you can set this to `view`.
- **MFT_AGENT_CONFIG_FILE** - Required. Path of the json file containing information required for setting up an agent. The path must be on a mount point. For example a configMap on OpenShift. See the [agent configuration doc](https://github.com/ibm-messaging/mft-cloud/blob/mftubi/docs//agentconfig.md) for a detailed description of attributes.
- **MFT_AGENT_NAME** - Required. Name of the agent to configure. 
- **BFG_JVM_PROPERTIES** - Optional - Any JVM property that needs to be set when running agent JVM.
- **MFT_LOG_LEVEL** - Optional - Defines the level of logging. `info` is default level of logging. `verbose` level displays more detailed logs.

The following command creates agent configuration and logs on the container file system. Configuration and logs will be deleted once the container ends.
`docker run --mount type=volume,source=mftagentcfg,target=/mftagentcfg --env LICENSE=accept --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/agentcfg.json  docker.io/ibmcom/mqmft:latest`

The following command creates agent configuration and logs on a mounted file system at `/mnt/mftadata` path. Hence configuration and logs will be available even after the container ends.
`docker volume create mftdata`   

`docker run --mount type=volume,source=mftagentcfg,target=/mftagentcfg --mount type=volume,source=mftdata,target=/mnt/mftdata -e MFT_AGENT_NAME=SRCAGENT --env LICENSE=accept --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/agentcfg.json  docker.io/ibmcom/mqmft:latest`
Note: The mounted path provided for agent configuration and logs must have read/write permissions for other users.

The following command uses a mounted file system for transferring files. 

`docker volume create customerdata`   

The `customerdata` is mounted into the container as `/mountpath` path. A mount point is not required for BRIDGE agents as they send/recive files to FTP/SFTP/FTPS server.
`docker run --mount type=volume,source=mftagentcfg,target=/mftagentcfg --mount type=volume,source=customerdata,target=/mountpath -e MFT_AGENT_NAME=SRCAGENT --env LICENSE=accept --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/agentcfg.json  docker.io/ibmcom/mqmft:latest`
Note: The mounted path provided must have read/write permissions for other users.

Use `docker ps` command view the status of container.

MFT commands now be executed using container shell. For example:
`docker exec <image id> bash -c 'fteListAgents'`

To login into terminal of container and execute MFT commands

`docker exec it <image id> bash`
