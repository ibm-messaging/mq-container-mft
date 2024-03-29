
# Running the image with Podman Runtime
 This document describes the steps for using IBM MQ Managed File Transfer container image.

1) Download the image from IBM Container Registry repository
`podman pull icrio.io/ibm-messaging/mqmft:latest`

2) Create a json file, like agentcfg.json containig the information required for creating an agent.  The attributes of the JSON file are described [here](https://github.com/ibm-messaging/mft-cloud/tree/9.2.2/docs/agentconfig.md). 
Sample agent configuration JSON is here:
```
{
   "coordinationQMgr":{
      "name":"MFTCORDQM",
      "host":"coordqm.ibm.com",
      "port":1414,
      "channel":"MFTCORDSVRCONN",
      "qmgrCredentials" : {
         "mqUserId":"mquser",
         "mqPassword":"cGFzc3cwcmQ="
      }
   },
   "commandQMgr":{
      "name":"MFTCMDQM",
      "host":"cmdqm.ibm.com",
      "port":1414,
      "channel":"MFTCMDSVRCONN",
      "qmgrCredentials" : {
         "mqUserId":"mquser",
         "mqPassword":"cGFzc3cwcmQ="
      },
      "additionalProperties" : {
      }
   },
   "agents":[{
      "name":"SRCAGENT",
      "type":"STANDARD",
      "qmgrName":"MFTAGENTQM",
      "qmgrHost":"agentqm.ibm.com",
      "qmgrPort":1414,
      "qmgrChannel":"MFTSVRCONN",
      "qmgrCredentials" : {
         "mqUserId":"mquser",
         "mqPassword":"cGFzc3cwcmQ="
      },
      "additionalProperties":{
         "enableQueueInputOutput":"true",
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
      "qmgrCredentials" : {
         "mqUserId":"mquser",
         "mqPassword":"cGFzc3cwcmQ="
      },
	  "additionalProperties": {
		 "protocolBridgeCredentialConfiguration" : "/mqmftbridgecred/agentcreds/ProtocolBridgeCredentials.prop"
      }
   }]
}
```

3) Create volume on host system. This volume needs to be mounted on a container when running.
`podman volume create mftagentcfg`   

4) Determine the path of the volume on local file system
`podman volume inspect mftagentcfg`
   The output would be something like
```
   [
    {
        "CreatedAt": "2021-03-11T03:11:18-07:00",
        "Driver": "local",
        "Labels": {},
        "Mountpoint": "/var/lib/containers/storage/volumes/mfagentcfg/_data",
        "Name": "mftdata",
        "Options": {},
        "Scope": "local"
    }
   ]
```

5) Create a sub-directory under the directory shown in Mountpoint attribute above.
`mkdir /var/lib/containers/storage/volumes/mfagentcfg/_data/agentcfg`
`cp agentcfg.json /var/lib/containers/storage/volumes/mfagentcfg/_data/agentcfg`

6) Create a MQMFTCredentials.xml file and put user mapping for the queue manager you are connecting to.
	For example: Note that "user" attribute has not specified in the entry. This allows agent to pick the credentials for any user that it is running under.
```
<?xml version="1.0" encoding="UTF-8"?>
<tns:mqmftCredentials xmlns:tns="http://wmqfte.ibm.com/MQMFTCredentials" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://wmqfte.ibm.com/MQMFTCredentials MQMFTCredentials.xsd">
    <tns:qmgr name="MFTQM" mqUserId="mftqmuserid" mqPassword="mftqmpassw0rd"/>
</tns:mqmftCredentials>
```
	
7) Run chmod on MQMFTCredentials.xml file:
`cp MQMFTCredentials.xml /var/lib/containers/storage/volumes/mfagentcfg/_data/agentcfg`

8) Run the container using podman run command.
  Environment variables to be passed the podman run command
- **LICENSE** - Required. Set this to `accept` to agree to the MQ Advanced for Developers license. If you wish to see the license you can set this to `view`.
- **MFT_AGENT_CONFIG_FILE** - Required. Path of the json file containing information required for setting up an agent. The path must be on a mount point. For example a configMap on OpenShift. See the [agent configuration doc](docs/agentconfig.md) for a detailed description of attributes.
- **MFT_AGENT_NAME** - Required. Name of the agent to configure. 
- **BFG_JVM_PROPERTIES** - Optional - Any JVM property that needs to be set when running agent JVM.
- **MFT_LOG_LEVEL** - Optional - Defines the level of logging. `info` is default level of logging. `verbose` level displays more detailed logs.

The following command creates agent configuration and logs on the container file system. Configuration and logs will be deleted once the container ends.
`podman run --mount type=volume,source=mftagentcfg,target=/mftagentcfg --env LICENSE=accept --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/agentcfg.json  icr.io/ibm-messaging/mqmft:latest`

The following command creates agent configuration and logs on a mounted file system at `/mnt/mftadata` path. Hence configuration and logs will be available even after the container ends.
`podman volume create mftdata`   

`podman run --mount type=volume,source=mftagentcfg,target=/mftagentcfg --mount type=volume,source=mftdata,target=/mnt/mftdata -e MFT_AGENT_NAME=SRCAGENT --env LICENSE=accept --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/agentcfg.json  icr.io/ibm-messaging/mqmft:latest`
Note: The mounted path provided for agent configuration and logs must have read/write permissions for other users.

The following command uses a mounted file system for transferring files. 

`podman volume create customerdata`   

The `customerdata` is mounted into the container as `/mountpath` path. A mount point is not required for BRIDGE agents as they send/recive files to FTP/SFTP/FTPS server.
`podman run --mount type=volume,source=mftagentcfg,target=/mftagentcfg --mount type=volume,source=customerdata,target=/mountpath -e MFT_AGENT_NAME=SRCAGENT --env LICENSE=accept --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/agentcfg.json  icr.io/ibm-messaging/mqmft:latest`
Note: The mounted path provided must have read/write permissions for other users.

Use `podman ps` command view the status of container.

MFT commands now be executed using container shell. For example:
`podman exec <image id> bash -c 'fteListAgents'`

To login into terminal of container and execute MFT commands

`podman exec it <image id> bash`
