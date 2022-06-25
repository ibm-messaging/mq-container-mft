
# Deploying agent image in OpenShift Container Platform

This document describes the steps for using IBM MQ Managed File Transfer (MFT) image in an OpenShift Container Platform. This document assumes that OpenShift Container Platform v4.6.6 or above is already setup and required CLI has been downloaded. The MFT container image will be pulled from IBM Container Registry for deployment.

1) Login to your OpenShift portal.

2) Get the login credentials by clicking on the "Copy login command" menu. Clicking on the menu will take you to another page where the login command will be displayed. 
	
	The command would look something like - 
	
	`oc login --token=<token> --server=<URL>`

3) Open a command prompt on your machine.

4) Run the command in step #2 to login to your OpenShift cluster.

	`oc login --token=<token> --server=<URL>`
	
5) Create a project using the command like below
	
	`oc new-project <project name>`

6) MFT requires MQ queue manager, so next step is to install IBM MQ. 

	1) An entitlement key is required. So head to https://www.ibm.com/support/knowledgecenter/SSGT7J_20.2/install/entitlement_key.html
	
	2) Copy the entitlement key and run the following command
	
	   `oc create secret docker-registry ibm-entitlement-key --docker-username=cp --docker-password=<entitlement key> --docker-server=cp.icr.io --namespace=<project name>`
	
	3) Next step is to install IBM MQ Operator and create a queue manager. So head to https://www.ibm.com/support/knowledgecenter/SSFKSJ_9.2.0/com.ibm.mq.ctr.doc/ctr_installing_ui.htm for instructions

7) Agent requires number of queue manager objects to be created before an agent can be configured and started. The following steps describe the required queue manager configuration. 

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
	
8) The next step is to configure an IBM MQ Managed File Transfer agent. 

   Required coordination, command and agent queue managers objects must be configured as described above.
	
   The information required to for the configuration must be provided as JSON data in a ConfigMap. The ConfigMap must be mounted in the agent container. The attributes of the JSON are described [here](https://github.com/ibm-messaging/mft-cloud/tree/9.2.2/docs/agentconfig.md).
	
	Create a yaml file with required attributes. Here is sample. Replace queue manager attributes with your queue manager attributes.
	
```
kind: ConfigMap
apiVersion: v1
metadata:
  name: mqmft-agent-config
  namespace: ibmmqmft
data:
  mqmftcfg.json: |
    {
      "coordinationQMgr":{
        "name":"MFTCORDQM",
        "host":"coordqm.ibm.com",
        "port":1414,
        "channel":"MFTCORDSVRCONN",
      "qmgrCredentials" : {
         "mqUserId":"mquser",
         "mqPassword":"cGFzc3cwcmQ="
      },
        "additionalProperties" : {
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
    }, {
      "name":"AGENTDEST",
      "type":"BRIDGE",
      "qmgrName":"MFTAGENTQM",
      "qmgrHost":"agentqmdest.mycomp.com",
      "qmgrPort":1818,
      "qmgrChannel":"MFTSVRCONN",
      "qmgrCredentials" : {
         "mqUserId":"mquser",
         "mqPassword":"cGFzc3cwcmQ="
      },
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
        "protocolBridgeCredentialConfiguration" : "/mnt/credentials/ProtocolBridgeCredentials.prop"
      }
    }] }

```

Run the following command to create the ConfigMap. Replace `configMap.yaml` with your file name.

`oc apply -f <configMap.yaml>`

   If you plan to deploy protocol bridge agent, then add required attributes to the configMap or create a separate configMap. Protocol bridge agent requires additional credentials to connect to FTP/FTPS/SFTP server. Those credentials can be provided as a key-value pair either through a Kubernetes ConfigMap or a Secret.
	
Description of credential file attributes are provided [here](https://github.com/ibm-messaging/mft-cloud/tree/9.2.2/config/ibm-mqmft-pba-creds-config.yaml). 

Sample configMap with plaintext password is here. 
```
kind: ConfigMap
apiVersion: v1
metadata:
  name: pba-custom-cred-map
  namespace: ibmmqmft
data:
  ProtocolBridgeCredentials.prop: sftp.server.com=root!0!passw0rd
```
Sample configMap with base64 encoded password is here. 

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: pba-custom-cred-map
  namespace: ibmmqmft
data:
  ProtocolBridgeCredentials.prop: sftp.server.com=root!1!S2l0dEBuMG9y
```
If you are providing bridge credentials through a Kubernetes secret, then the value must be base64 encoded. A sample is here.
```
kind: Secret
apiVersion: v1
metadata:
  name: pba-custom-cred-map
  namespace: ibmmqmft
data:
  ProtocolBridgeCredentials.prop: c2Z0cC5zZXJ2ZXIuY29tPXJvb3QhMCFwYXNzdzByZA==
```

Run the following command to create the ConfigMap. Replace `pba.yaml` with your file name.
	
`oc apply -f <pba.yaml>`

9) Starting with 9.2.4 CD release, MQ Managed File Transfer creates additional logs that tracks the progress of a transfer more granularly. The logs, in json format, are written to transferlog0.json file under the agent's log directory. The MFT container image provides an option using which transfer logs can be automatically pushed to logDNA server. The details of the logDNA server like the URL, injestion key must be provided as a JSON object through a kubernetes secret. The contents of the secret must be base64 encoded. Here is the JSON structure:

```
{ 
	"type":"logDNA", 
	"logDNA":{
		"url":"https://<your logdna host name>/logs/ingest",
		"injestionKey":"<your injestion key>"
	}
}
```

The entire content must be base64 encoded before putting into a secret. Then the secret must be mounted into the container. An example secret is here:

```
  kind: Secret
  apiVersion: v1
  metadata:
    name: logdna-secret
    namespace: ibmmqft
  data:
    logdna.json: >-
      <base64 encoded data>
```


10) Next step is to deploy the container image from IBM Container Registry or other image repository.

	Create deployment yaml for IBM MQ Managed File Transfer image. Refer the documentation [here](https://github.com/ibm-messaging/mft-cloud/tree/9.2.2/config/ibm-mqmft-deployment.yaml) for more details.
	
	Sample deployment yaml is described here.

```
kind: Deployment
apiVersion: apps/v1
metadata:
  name: ibm-mq-managed-file-transfer-srcagent
  namespace: ibmmqmft
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ibm-mq-managed-file-transfer-srcagent
  template:
    metadata:
      labels:
        app: ibm-mq-managed-file-transfer-srcagent
        deploymentconfig: ibm-mq-managed-file-transfer-srcagent
    spec:
      volumes:
        - name: mqmft-agent-config-map
          configMap:
            name: mqmft-agent-config
            defaultMode: 420
        - name: logdna-url-secret
          secret:
            secretName: logdna-secret
            items:
            - key: logdna.json
              path: logdna.json
      containers:
        - resources: {}
          readinessProbe:
            exec:
              command:
                - agentready
            initialDelaySeconds: 15
            timeoutSeconds: 3
            periodSeconds: 30
            successThreshold: 1
            failureThreshold: 3
          terminationMessagePath: /mnt/termination-log
          name: ibm-mq-managed-file-transfer-srcagent
          livenessProbe:
            exec:
              command:
                - agentalive
            initialDelaySeconds: 90
            timeoutSeconds: 5
            periodSeconds: 90
            successThreshold: 1
            failureThreshold: 3
          env:
            - name: MFT_AGENT_NAME
              value: SRCAGENT
            - name: LICENSE
              value: accept
            - name: MFT_AGENT_CONFIG_FILE
              value: /mqmftcfg/agentconfig/mqmftcfg.json
            - name: MFT_LOG_LEVEL
              value: "info"
            - name: MFT_AGENT_TRANSFER_LOG_PUBLISH_CONFIG_FILE
              value: /logdna/logdna.json
          imagePullPolicy: Always
          volumeMounts:
            - name: mqmft-agent-config-map
              mountPath: /mqmftcfg/agentconfig
            - name: logdna-url-secret
              mountPath: /logdna/logdna.json
              subPath: logdna.json
              readOnly: true              
          terminationMessagePolicy: File
          image: >-
             icr.io/ibm-messaging/mqmft:latest
      restartPolicy: Always

```
Then run the following command to deploy the image. Replace `deployment.yaml` with your file name

`oc apply -f <deployment.yaml>`
		
11) Login into your OCP Cluster portal and verify the agent is running or use the following commands to verify

	`oc get pods`
	
	`oc describe pod <pod name>`

12) Login to the terminal of the pod and do any further configuration like creating resource monitor, scheduled transfer or submitting transfers can be done via terminal. 
	
	Run the following command to display status of your agent
	
	`fteShowAgentDetails <agent name>`
	
	Run the following command to display status of all agents under the coordination queue manager

	`fteListAgents <agent name>`

13) Initiating transfers

	Agent will have a restricted access to the file system of the container. Agent can read from or write to only the `/mountpath` folder on the file system and not any other part of the file system. This folder can also be mapped to a external file system via a mount point. 
	
	The mount point can be specified through the deplyment yaml like 

```
	spec:
      volumes:
        - name: mqmft-nfs
          persistentVolumeClaim:
            claimName: nfs-pvc
      containers:
          env:
            - name: MFT_MOUNT_PATH
              value: /mountpath
	      volumeMounts:
            - name: mqmft-nfs
              mountPath: /mountpath
```

You can use fteCreateTransfer command to initiate a transfer. 

You can also setup resource monitors using fteCreateMonitor command to trigger transfers.

### IMPORTANT NOTE: 
1) You may need to manually escape `$` characters before issuing the fteCreateTransfer command otherwise Linux shell can incorrectly interpret the `$`. See example below
2) Since the user will have access only to `/mountpath`, any MFT command that creates a file, must use `/mountpath` as the target directory. 

For example the following command creates a task.xml under `/mountpath/mntr` directory. 
`fteCreateTransfer -gt /mountpath/mntr/task.xml -sa ATCFG -sm QUICKSTART -da BRIDGE -dm QUICKSTART -de overwrite -df "sftp://10.17.68.52/\${FileName}" "\${FilePath}"`

`fteCreateMonitor -ma ATCFG -mn F2B -md "/mountpath" -tr "match,*.txt" -f -mt task.xml`


### Viewing status of transfers.

You can configure MQExplorer MFT Plugin on-premise to monitor status of transfers and other MFT resources. Follow the steps here to [here](https://github.com/ibm-messaging/mft-cloud/tree/9.2.2/config/connectmqexplorer.md) to configure MQExplorer on premise to connect to queue manager on OpenShift cluster.

You can also view the status of transfers by running `mqfts` command on the terminal of your pod. The `mqfts` command lists the status of transfer by parsing capture0.log file of the agent running in the pod. This means you can view the status of transfers where the agent in the current pod is a source agent.

Example:

`mqfts` - Lists status of transfers of this agent.

`mqfts --id=414D51204D4654514D2020202020202038EE4560223DE303` - Displays details of single transfer.

`mqfts --h` - Displays help

`mqfts --lf <path to captureX.log file>` Display transfer status from the specified file. `X` represents a number.

   For example: 
	 
	mqfts --lf /mnt/mftdata/mqft/logs/QMC/agents/SRC/logs/capure1.log
	mqfts --lf=/mnt/mftdata/mqft/logs/QMC/agents/SRC/logs/capure1.log --id=414D51204D4654514D2020202020202038EE4560223DE303'