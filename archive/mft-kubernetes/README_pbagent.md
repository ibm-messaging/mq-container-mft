---
copyright:
  years: 2017, 2020
lastupdated: "2020-05-19"
---

# MQ Managed File Transfer for Kubernetes

### Protocol Bridge Agent setup for MFT Kubernetes cluster

2. We will create a source agent on kubernetes cluster as part of this tutorial. These agents are **AGENTSRC**. As a first step of agent configuration, we have to create their congfiguration on the coordination queue manager.

    ```
    kubectl exec -ti <podname> /etc/mqm/mft/mqft_setupAgent.sh AGENTPBA
    ```
### MFT Agents deployment on to Kubernetes cluster

Create a new deployment file(mft_agentredit_Deployment-pba-agent.yaml) for deploying the MFT Agent as a pod and container within it.

```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: mft-kube-deployment-src
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mqft-kube-src
    spec:
      containers:
      - name: agentsrc
        image: registry.eu-de.bluemix.net/mft-images/agentredist_mftimagerepo:1.0
        env:
        - name: MQ_QMGR_NAME
          value: "QM1"
        - name: MQ_QMGR_HOST
          value: "169.51.10.240"
        - name: MQ_QMGR_PORT
          value: "32075"
        - name: MQ_QMGR_CHL
          value: "MFT.SVRCONN"
        - name: MFT_AGENT_NAME
          value: "AGENTSRC"
        - name: IS_PBA_AGENT
          value: "true"
        - name: PROTOCOL_FILE_SERVER_TYPE
          value: <userinputvalues>
        - name: SERVER_HOST_NAME
          value: <userinputvalues>
        - name: SERVER_TIME_ZONE
          value: <userinputvalues>
        - name: SERVER_PLATFORM
          value: <userinputvalues>
        - name: SERVER_LOCALE
          value: <userinputvalues>
        - name: SERVER_FILE_ENCODING
          value: <userinputvalues>
```

  **Note:**     
  1. the \"<userinputvalues\>" are the configuration values required for setting up the PBA Agent. For more detailed information [click here](https://www.ibm.com/support/knowledgecenter/SSFKSJ_9.1.0/com.ibm.mq.ref.adm.doc/create_bridge_agent_cmd.htm) .

### Apply the deployment on your Kubernetes cluster

1. Run following command to create the MFT Agent deployment
    ```
    kubectl create -f filepath/mft_agentredit_Deployment-pba-agent.yaml
    ```
2. Check if the deployment is successful and pod is in running state. Run the below command with few seconds gap until you find the pod status to be running
    ```
    kubectl get pods
    ```
3. If incase pods goes into error state. Review the logs to investigate and debug on error cause.
    ```
    kubectl describe pod <pod-name>
    kubectl logs <pod-name> -c <container-name>
    ```
    **Note:** 
    1. `<pod-name>` can be found in the output of command `kubectl get pods`
    2. `<container-name>` can be found in the mqadvdeployment.yaml file(at spec-->containers-->name)
    3. Above command creates a template file **ProtocolBridgeProperties.xml**, in the agent configuration directory MQ_DATA_PATH/mqft/config/coordination_queue_manager/agents/bridge_agent_name. The command also creates an entry in the file for the default protocol file server, if a default was specified when the command was run.
    4. To Specify more than one server configuration in the ProtocolBridgeProperties.xml file you will have to use kubectl exec into the pba agent pod and modify the file and then save copy of it for the further use.