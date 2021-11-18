This document describes the steps to setup an IBM MQ Managed File Transfer agent with NFS share mounted. NFS Server will be configured on the infrastructure node of a OpenShift cluster. The NFS server will use the local file system of the infrastructure node. A directory of the local file system will be mounted as NFS share into IBM MQ Managed File Transfer container.

### Basic requirements:
1) OpenShift CLI - `oc` command. This command will already be installed on the infrastructure node.

### Setup NFS server on infrastructure node of OpenShift Cluster
The first step is to setup NFS Server on infrastructure node:

1) SSH to your infrastructure node with user id and password provided by OpenShift.

2) Run `ip addr show` command and note down the internal IP address of the node. Let's assume the IP address is `10.17.12.68`

3) Install and Configure NFS

	`sudo dnf install nfs-utils -y`

4) Once the installation is complete, start and enable the nfs-server. Run the following commands:

	`sudo systemctl start nfs-server.service`
	
	`sudo systemctl enable nfs-server.service`

5) To confirm that NFS service is running, execute:

	`sudo systemctl status nfs-server.service`

6) You can verify the version of nfs protocol that you are running by executing the command:

	`rpcinfo -p | grep nfs`

7) Create a NFS share and exporting it

	`sudo mkdir -p /mnt/nfs_shares/mqft`

8) To avoid file restrictions on the NFS share directory, it’s advisable to configure directory ownership as shown. This allows MFT agents to create files without encountering any permission issues.

	`sudo chown -R nobody: /mnt/nfs_shares/mqft`

9) Also, you can decide to adjust the directory permissions according to your preference. For instance, in this guide, we will assign all the permissions (read , write and execute) to the NFS share folder

	`sudo chmod -R 777 /mnt/nfs_shares/mqft`
	
10) For the changes to come into effect, restart the NFS daemon:

	`sudo systemctl restart nfs-utils.service`

11) Export NFS share for clients to access.  Add the following line in the file `/etc/exports`:

	`sudo vi /etc/exports`
	
	`/mnt/nfs_shares/mqft    10.17.*(rw,sync,no_all_squash,root_squash)`

12) Verify the contents by running following command:
			
		cat /etc/exports
	

13) To export the above created folder, use the exportfs command as shown:

	`sudo exportfs -arv`
	
This completes setting up of NFS Server on infrastructure node. The next step is to create a persistent volume and claims that uses NFS share created above.

### Creating PersistentVolume and PersistentVolumeClaim
1) Login to OpenShift cluster with the following command

`oc login --token=<replace with your token\> --server=<replace with your cluster URL\>`

2) Create a persistent volume with /mnt/nfs_shares/mqft directory on infrastructure node. Create a yaml file say `pv.yaml`, with the following content. 
```
apiVersion: v1
kind: PersistentVolume
metadata:
  name: mqftnfs-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  nfs:
    path: /mnt/nfs_shares/mqft
    server: 10.17.12.68
```
Run the following command to create a volume:

`oc apply -f pv.yaml`

To verify the volume creation, run the following command:

`oc get volumes`

2) Create a persistent claim. Create a yaml file, say `pvc.yaml` with following content. 
```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mqftnfs-pv-claim
  namespace: mqmft
spec:
  storageClassName: manual
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 5Gi
```	

This completes setting up of a Persistent volume and a Persistent volume claim. 

### Creating pull secret
The next step is to create a secret to pull IBM MQ Managed File Transfer container image from a private repository. This step is not required if you are deploying the container from dockerHub.

Run the following command to create a pull secret. Replace the command parameters with your repository details.
```
oc create secret docker-registry mqmftpullsecret \
    --docker-server=<your private repository server URI\> \
    --docker-username=<UID of repository\> \
    --docker-password=<Password for UID\> \
    --docker-email=<email id \>
 ```

Make the secret for image pull.

`oc secrets link default mqmftpullsecret --for=pull`

`oc secrets link builder mqmftpullsecret `

### Create image pull secret for IBM MQ
Follow the steps provided in the link below to create an image pull secret for IBM MQ.
https://www.ibm.com/docs/en/ibm-mq/9.2?topic=pyopm-preparing-your-openshift-project-mq-using-openshift-web-console

### Installing IBM MQ Operator
Once the image pull secret for IBM MQ is created, the next step is to deploy IBM MQ operator.

1) Create a yaml, say `csoperator.yaml` file with following content.
```
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: opencloud-operators
  namespace: openshift-marketplace
spec:
  displayName: IBMCS Operators
  publisher: IBM
  sourceType: grpc
  image: icr.io/cpopen/ibm-common-service-catalog:latest
  updateStrategy:
    registryPoll:
      interval: 45m
```
Run the following command the deploy the `common services` operator.

`oc apply -f csoperator.yaml`

Then create IBM operator using the following yaml. Save the following content to a yaml file, say `ibmoperator.yaml`.

```
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: ibm-operator-catalog
  namespace: openshift-marketplace
spec:
  displayName: IBM Operator Catalog
  image: icr.io/cpopen/ibm-operator-catalog:latest
  publisher: IBM
  sourceType: grpc
  updateStrategy:
    registryPoll:
      interval: 45m
```

Run the following command to deploy the IBM operator:
`oc apply -f ibmoperator.yaml`

### Deploy queue manager 
Once the operator is deployed, the next step is to create a config map containing MQSC that defines objects required for IBM MQ Managed File Transfer.
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: mqsc-mqmft-example
data:
  mqmftcoord.mqsc: |
    DEFINE CHL (QS_SVRCONN) CHLTYPE (SVRCONN)
	ALTER QMGR CHLAUTH(DISABLED)
	DEFINE TOPIC('SYSTEM.FTE') TOPICSTR('SYSTEM.FTE') REPLACE
	ALTER TOPIC('SYSTEM.FTE') NPMSGDLV(ALLAVAIL) PMSGDLV(ALLAVAIL)
	DEFINE QLOCAL(SYSTEM.FTE) LIKE(SYSTEM.BROKER.DEFAULT.STREAM) REPLACE
	ALTER QLOCAL(SYSTEM.FTE) DESCR('Stream for MQMFT Pub/Sub interface')
	DISPLAY NAMELIST(SYSTEM.QPUBSUB.QUEUE.NAMELIST)
	ALTER NAMELIST(SYSTEM.QPUBSUB.QUEUE.NAMELIST) NAMES(SYSTEM.BROKER.DEFAULT.STREAM,SYSTEM.BROKER.ADMIN.STREAM,SYSTEM.FTE)
	DISPLAY QMGR PSMODE
	ALTER QMGR PSMODE(ENABLED)
  mqmftagent.mqsc: |
    DEFINE CHANNEL(AGENTQMCHL) CHLTYPE(SVRCONN) TRPTYPE(TCP) SSLCAUTH(OPTIONAL) SSLCIPH('ANY_TLS12_OR_HIGHER')
    SET CHLAUTH(AGENTQMCHL) TYPE(BLOCKUSER) USERLIST('nobody') ACTION(ADD)
    DEFINE QLOCAL(SYSTEM.FTE.COMMAND.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.DATA.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.REPLY.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.STATE.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.EVENT.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.AUTHAGT1.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.AUTHTRN1.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.AUTHOPS1.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.AUTHSCH1.SRCSTD) +
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
        DEFINE QLOCAL(SYSTEM.FTE.AUTHMON1.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.AUTHADM1.SRCSTD) +
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
    DEFINE QLOCAL(SYSTEM.FTE.HA.SRCSTD) +
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

Refer the following pages for more details.

https://www.ibm.com/docs/en/ibm-mq/9.2?topic=umicpio-deploying-configuring-mq-certified-containers-using-mq-operator
https://www.ibm.com/docs/en/ibm-mq/9.2?topic=manager-example-supplying-mqsc-ini-files

The next step is to deploy the queue manager. Create a yaml file, say `qm.yaml` with the following contents.
```
apiVersion: mq.ibm.com/v1beta1
kind: QueueManager
metadata:
  name: mqmftqm
spec:
  version: 9.2.3.0-r1
  license:
    accept: true
    license: L-APIG-BMJJBM
  web:
    enabled: true
  queueManager:
    name: "QUICKSTART"
	mqsc:
    - configMap:
        name: mqsc-ini-example
        items:
        - mqmftagent.mqsc
        - mqmftcoord.mqsc
	storage:
      queueManager:
        type: ephemeral
  template:
    pod:
      containers:
       - name: qmgr
         env:
         - name: MQSNOAUT
           value: "yes"
```
Run the following command to deploy queue manager:

`oc apply -f qm.yaml`

Wait for queue manager to start. Once the queue manager pods starts, note down the IP address of the pod. Let's assume the IP address of the queue manager pod as `10.17.28.28`.

Once the queue manager starts, the next step is to deploy agent.

### Deploy an agent 
Before we deploy the agent, we need to define the attributes for the agent. Create a yaml file, say `agentconfig.yaml` with the following contents. The yaml defines attributes for coordination queue manager, command queue manager and agent.
```
kind: ConfigMap
apiVersion: v1
metadata:
  name: mqmft-agent-config
  namespace: ibmmqmft
data:
  mqmftcfg.json: >
    {"waitTimeToStart":20,"coordinationQMgr":{"name":"QUICKSTART","host":"10.17.28.28","port":1414,"channel":"QS_SVRCONN","additionalProperties":{}},"commandQMgr":{"name":"QUICKSTART","host":"10.17.28.28","port":1414,"channel":"QS_SVRCONN","additionalProperties": {}},"agents":[{"name":"SRCSTD","deleteOnTermination":"true","cleanOnStart":"all","type":"STANDARD","qmgrName":"QUICKSTART","qmgrHost":"10.17.28.28","qmgrPort":1414,"qmgrChannel":"QS_SVRCONN","additionalProperties":{"enableQueueInputOutput":"true",     "trace":"all","logCapture":"true"}}]}
```
Run the following command to create the configMap:
`oc apply -f agentconfig.yaml`.

The next step is to deploy the agent. Create a yaml file, say `agentdep.yaml` with the following contents. Notice the Persistent Volume, Persistent Volume Claim and the mount point into the container.
Also let's assume private repository URI from where agent image will be pulled as `myimagerep.mycomp.com/ibmmqmqft:latest`.
```
kind: Deployment
apiVersion: apps/v1
metadata:
  name: ibm-mq-managed-file-transfer-srcstd
  namespace: mqmft
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ibm-mq-managed-file-transfer-srcstd
  template:
    metadata:
      labels:
        app: ibm-mq-managed-file-transfer-srcstd
        deploymentconfig: ibm-mq-managed-file-transfer-srcstd
    spec:
      volumes:
        - name: mqmft-agent-config-map
          configMap:
            name: mqmft-agent-config
            defaultMode: 420
        - name: mqftnfs-pv
          persistentVolumeClaim:
            claimName: 
              mqftnfs-pv-claim
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
          terminationMessagePath: /mountpath/termination-log
          name: ibm-mq-managed-file-transfer-srcstd
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
              value: SRCSTD
            - name: LICENSE
              value: accept
            - name: MFT_AGENT_CONFIG_FILE
              value: /mqmftcfg/agentconfig/mqmftcfg.json
          imagePullPolicy: Always
          volumeMounts:
            - name: mqmft-agent-config-map
              mountPath: /mqmftcfg/agentconfig
            - name: mqftnfs-pv
              mountPath: /mountpath
          terminationMessagePolicy: File
          image: >-
             myimagerep.mycomp.com/ibmmqmqft:latest
      restartPolicy: Always
      terminationGracePeriodSeconds: 60
      dnsPolicy: ClusterFirst
      securityContext: {}
      schedulerName: default-scheduler
```

Then run the following command to deploy the agent:

`oc apply -f agentdep.yaml`

View the events and logs on the OpenShift console for any errors. If all goes well, the agent pod will start and run the agent. You can login to the terminal of the pod and run MFT commands.


After logging into MFT agent pod terminal, run the following command to create a sample file.

`echo "Hello MFT agent - writing file to NFS share" > /mountpath/input/hellonfs.txt` 

Then initiate a transfer by running the following command.
`fteCreateTransfer -rt -1 -sa SRCSTD -sm QUICKSTART -da SRCSTD -dm QUICKSTART -de overwrite -df /mntpath/output/hellonfs.txt /mntpath/input/hellonfs.txt`

NOTE: Agent can read/write files from/to to `/mountpath` directory only as the agent has been sandboxed to this directory.

Once the command finishes, verify the transfer result by running the following command. 

`mqfts`

If the result is `Failed`, run the following command to view more details:

`mqfts --id=<transfer id>`

You can also run `ls` command on `/mountpath/output` directory to verify if the file has been created or not.
