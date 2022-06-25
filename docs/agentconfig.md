# Agent configuration file
Agent is created and started during container creation time. The information required for creation of agent, like the agent name, coordination queue manager, agent queue manager etc must be provided via a json file located on a mount point. The path of the json file must be passed as a value to **MFT_AGENT_CONFIG_FILE** environment variable. 

The configuration file can contain attributes for multiple agents. However all agents will be created under the same cooridation queue manager.

This document describes attributes of the json file.

- **coordinationQMgr** - Type: Group. Defines the configuration information coordination queue manager.
- **name** - Type: String. Name of the coordination queue manager.
- **host** - Type: String. Host name to be used for connecting to coordination queue manager.
- **port** - Type: int. Port number to be used for connecting to coordination queue manager.
- **channel** - Type: String. Channel name to be used for connecting to coordination queue manager.
- **qmgrCredentials** - Type: Group. Defines the credentials required for connecting to coordination queue manager. The credentials provided here are save to MQMFTCredentials.xml file during container start.
- **mqUserId** - Type: String. Name of user for connecting to coordination queue manager.
- **mqPassword** - Type: String. Password of user for connecting to coordination queue manager. Recommended to base64 encode this value.
- **additionalProperties** - Optional. Type: Group. Any additional parameters to be set in coordination.properties file of the container. Names of the attributes in this group must match the name of properties in coordination.properties file.

- **commandQMgr** - Type: Group. Defines the configuration information for a command queue manager.
- **name** - Type: String. Name of the command queue manager.
- **host** - Type: String. Host name to be used for connecting to command queue manager.
- **port** - Type: int. Port number to be used for connecting to command queue manager.
- **channel** - Type: String. Channel name to be used for connecting to command queue manager.
- **qmgrCredentials** - Type: Group. Defines the credentials required for connecting to coordination queue manager. The credentials provided here are save to MQMFTCredentials.xml file during container start.
- **mqUserId** - Type: String. Name of user for connecting to command queue manager.
- **mqPassword** - Type: String. Password of user for connecting to command queue manager. Recommended to base64 encode this value.
- **additionalProperties** - Optional. Type: Group. Any additional parameters to be set in command.properties file of the container. Name of the attribute in this group must match the name of properties in command.properties file.

- **agents** - Type: Group. Defines an array of configuration information of agent. You can define multiple agent configuration. This allows same JSON file to be used for creating multiple agents. All agents would use the same coordination and command queue managers.
- **name** - Type: String. Name of the agent to be created.
- **type** - Optional. Type: String. Type of the agent to be created. `STANDARD` and `BRIDGE` are the supported values. Default is `STANDARD`.
- **cleanOnStart** - Optional. Type: String. Delete all pending transfers, resource monitors, any invalid messages, scheduled transfers. Supported values - one of `transfers`, `monitors`, `scheduledTransfers`, `invalidMessages` or `all`
- **deleteOnTermination** - Optional. Type: String. Deletes and deregisters an agent when a container ends. Supported values - `true` and `false`. Default is `false`.
- **qmgrName** - Type: String. Name of the queue manager to which agent will connect.
- **qmgrHost** - Type: String. Host name to be used for connecting to agent queue manager.
- **qmgrPort** - Type: int. Port number to be used for connecting to agent queue manager.
- **qmgrChannel** - Type: String. Channel name to be used for connecting to agent queue manager.
- **qmgrCredentials** - Type: Group. Defines the credentials required for connecting to coordination queue manager. The credentials provided here are save to MQMFTCredentials.xml file during container start.
- **mqUserId** - Type: String. Name of user for connecting to agent queue manager.
- **mqPassword** - Type: String. Password of user for connecting to agent queue manager. Recommended to base64 encode this value.
- **additionalProperties** - Type: Group. Any additional parameters to be set in agent.properties file of the container. Name of the attribute in this group must match the name of properties in agent.properties file.
- **protocolBridgeCredentialConfiguration** Type: String. Path of the custom protocol bridge credential file. This property must be set if the agent is of type BRIDGE. This file must contain "key=value" pair(s) containing credential information.
- **protocolBridge** - Required for BRIDGE agent. Type: JSONArray. Contains group of elements that defines additional properties if the agent type is `BRIDGE`.
- **serverType** - Type: String. Defines the protocol bridge type. `FTP`, `FTPS` and `SFTP` are the supported types.
- **serverHost** Type: String. Host name of the protocol server the agent will connect to. 
- **serverTimezone** Type: String. Timezone where protocol server is running. For example `Europe/London`. Valid only if serverType is `FTP` or `FTPS`.
- **serverPlatform** Type: String. Name of the platform on which protocol server is running. For example `UNIX`.
- **serverLocale** Type: String. Locale of the machine where protocol server is running. For example `en-GB`
- **serverListFormat** Type: String. Directory listing format of protocol server. For example `UNIX` or `Windows` os `OS400IFS`.
- **serverLimitedWrite** Type: String. Is server a limited function type. 
- **serverFileEncoding** Type: String. File encoding, for example `UTF8`

An example json is here:

```
{
   "coordinationQMgr":{
      "name":"MFTCORDQM",
      "host":"coordqm.ibm.com",
      "port":1414,
      "channel":"MFT_CORD_CHN",
      "qmgrCredentials" : {
         "mqUserId":"JohnDover",
         "mqPassword":"bXlwYXNzdzByZA==",
      },
      "additionalProperties" : {
         "coordinationQMgrStandby":"9.20.20.20(1414)"
      }
   },
   "commandsQMgr":{
      "name":"MFTCMDQM",
      "host":"cmdqm.ibm.com",
      "port":1414,
      "channel":"MFT_CMD_CHN",
      "additionalProperties" : {
         "connectionQMgrStandby":"9.20.20.20(1414)"
      },
      "qmgrCredentials" : {
         "mqUserId":"JohnDover",
         "mqPassword":"bXlwYXNzdzByZA==",
      },
   },
   "agents":[{
      "name":"AGENTSRC",
      "type":"STANDARD",
      "qmgrName":"MFTAGENTQM",
      "qmgrHost":"agentqm.ibm.com",
      "qmgrPort":1414,
      "qmgrChannel":"MFT_AGENT_CHN",
      "additionalProperties":{
         "enableQueueInputOutput":"true"
      },
      "qmgrCredentials" : {
         "mqUserId":"JohnDover",
         "mqPassword":"bXlwYXNzdzByZA==",
      },
   },
   {
      "name":"AGENTDEST",
      "type":"BRIDGE",
      "qmgrName":"MFTAGENTDESTQM",
      "qmgrHost":"agentqmdest.mycomp.com",
      "qmgrPort":1818,
      "qmgrChannel":"MFT_AGENT_CHN",
      "qmgrCredentials" : {
         "mqUserId":"JohnDover",
         "mqPassword":"bXlwYXNzdzByZA==",
      },
      "protocolBridge" : [{
         "serverType":"FTP",
         "serverHost":"ftp.mycomp.com",
         "serverTimezone":"Europe/London",
         "serverPlatform":"UNIX",
         "serverLocale":"en-GB",
         "serverListFormat":"UNIX", 
         "serverLimitedWrite":"false", 
         "serverFileEncoding":"UTF8", 
         "serverPassiveMode":"true", 
	  }],
	  "additionalProperties": {
		 "protocolBridgeCredentialConfiguration" : "/mqmftbridgecred/agentcreds/ProtocolBridgeCredentials.prop"
      },
   }]
}
```