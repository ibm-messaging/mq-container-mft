---
copyright:
  years: 2017, 2020
lastupdated: "2020-06-09"
---

# MQ Managed File Transfer for Kubernetes

### Protocol Bridge Agent setup for MFT Kubernetes cluster

1. We will create a pba agent on kubernetes cluster as part of this tutorial. This agent is names **PBA**. As a first step of agent configuration, we have to create their congfiguration on the coordination queue manager.

    ```
    oc exec -ti <qmgr-pod-name> /etc/mqm/QMgrSetup/mqft_setupAgent.sh PBA
    ```
### MFT Agents deployment on to Kubernetes cluster

1. Create a new deployment file(mft_agentredit_Deployment-pba-agent.yaml) for deploying the MFT Agent as a pod and container within it.
 
**Note:** Currently the sample only enables you to configure it against a SFTP Server.
```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: mft-cp4i-pba
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mqft-cp4i
    spec:
      containers:
      - name: agentpba
        image: image-registry.openshift-image-registry.svc:5000/mq915mft/mftpbaagent:1.0
        imagePullPolicy: Always
        env:
        - name: MQ_QMGR_NAME
          value: "mqmft0806"
        - name: MQ_QMGR_HOST
          value: "10.254.18.54"
        - name: MQ_QMGR_PORT
          value: "1414"
        - name: MQ_QMGR_CHL
          value: "DEV.ADMIN.SVRCONN"
        - name: MFT_AGENT_NAME
          value: "PBA"
        - name: IS_PBA_AGENT
          value: "true"
        - name: PROTOCOL_FILE_SERVER_TYPE # Required if IS_PBA_AGENT=true e.g.: FTP,SFTP,FTPS
          value: "SFTP"
        - name: SERVER_HOST_NAME # Required if IS_PBA_AGENT=true
          value: "10.254.18.31"
        - name: SERVER_PLATFORM # UNIX, WINDOWS - Required if IS_PBA_AGENT=true
          value: "UNIX"
        - name: SERVER_FILE_ENCODING # Required if IS_PBA_AGENT=true e.g.: UTF-8
          value: "UTF-8"
```

  **Note:**     
  1. For more detailed information on PBA [click here](https://www.ibm.com/support/knowledgecenter/SSFKSJ_9.1.0/com.ibm.mq.ref.adm.doc/create_bridge_agent_cmd.htm).


### Apply the deployment on your Kubernetes cluster

1. Run following command to create the MFT Agent deployment
    ```
    oc create -f filepath/mft_agentredit_Deployment-pba-agent.yaml
    ```
2. Check if the deployment is successful and pod is in running state. Run the below command with few seconds gap until you find the pod status to be running
    ```
    oc get pods
    ```
3. If incase pods goes into error state. Review the logs to investigate and debug on error cause.
    ```
    oc describe pod <pod-name>
    oc logs <pod-name> -c <container-name>
    ```
    **Note:** 
    1. `<pod-name>` can be found in the output of command `kubectl get pods`
    2. `<container-name>` can be found in the mft_agentredit_Deployment-pba-agent.yaml file(at spec-->containers-->name)
    3. Above command creates a template file **ProtocolBridgeProperties.xml**, in the agent configuration directory MQ_DATA_PATH/mqft/config/coordination_queue_manager/agents/bridge_agent_name. The command also creates an entry in the file for the default protocol file server, if a default was specified when the command was run.
    4. To Specify more than one server configuration in the ProtocolBridgeProperties.xml file you will have to use `oc exec` into the pba agent pod and modify the file and then save copy of it for the further use.