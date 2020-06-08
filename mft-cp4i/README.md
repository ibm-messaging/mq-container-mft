**

## MFT Agents for Cloudpak for Integration platform

This tutorial will guide you in configuring *MFT Agents* on Cloudpak for integration platform. Further, this tutorial will guide you in performing MFT operations on the *Agents* such as listing agents, performing file transfers using the agents.  

**Prerequisites**:
 1. This tutorial assumes that you have created an Openshift CP4I cluster and having an IBM MQ Queue manager running under a cluster namespace. For more details on this  please refer - https://www.ibm.com/support/knowledgecenter/SSGT7J_20.1/welcome.html  
 2. Before you start on any steps in this tutorial, open a terminal and log into your Openshift cluster using `oc login` Here after we call such terminal (oc logged-in) as **oc-terminal**   
 3. This tutorial is using single queue manager topology for MFT configuration, where the same queue manager serves as Coordination, Command and Agent queue managers.  

### Tutorial Overview
1. Downloading mft-cp4i package and cluster setup  
2. Setup queue manager as Coordination queue manager  
3. Setup queue manager Agent queue manager  
4. Build MFT Agent Docker image
5. Push MFT Agent Docker image to your cluster namespace  
6. Deploy MFT Agent - A1
7. Deploy MFT Agent - A2
8. List Agents
9. Initial a file transfer from A1⇒ A2
10. Verify file transfer was successful and file exits on A2 container.

#### 1. Downloading mft-cp4i package and cluster setup

 1. Download/clone the mft-cp4i package from GitHub   
	(https://github.com/ibm-messaging/mft-cloud/tree/master/mft-cp4i)  
 2.  Perform all the below steps on oc-terminal   
 3. Navigate to `<path>/mft-cp4i`  
 4. Copy directory *QMgrSetup* into your already running qmgr container. For ex(update qmgr-pod-name as per your setup):  
	 `kubectl cp <path>/mft-cp4i/QMgrSetup <qmgr-pod-name>:/etc/mqm/`  
 5. Give +x permission to the script files. For ex (update qmgr-pod-name as per your setup):  
	 ```
	 oc exec -ti <qmgr-pod-name> chmod +x /etc/mqm/QMgrSetup/mqft_setupCoordination.sh
	 oc exec -ti <qmgr-pod-name> chmod +x /etc/mqm/QMgrSetup/mqft_setupAgent.sh
	 ```   
	 i. *mft_setupCoordination.sh* sets up a queue manager as the coordination queue manager. Script will create all the required objects by mft coordination queue manager. Hence, this script is required to run only for coordination queue manager setup.  
	 ii. *mft_setupAgent.sh* sets up a queue manager as the agent queue manager.  Script will create all the required objects by mft agent queue manager. Hence, this script is required to run only for agent queue manager setup.  
		 
 6. Run the steps 3-5 on all the queue managers that are coordination/agent queue managers.  
 7. By default, MFT Agent Dockerfile creates mftadmin user to handle all the mft operations within the container. Hence, below command has to be run to allow user to have access to perform mft operations.  
 ``oc adm policy add-scc-to-group anyuid system:serviceaccounts:<namespace>``  

#### 2. Setting up Queue manager as Coordination Queue manager  

 1. `mqft_setupCoordination.sh` script is provided to automate the creation of all mft objects required on a queue manager for it to serve as coordination queue manager.  
 2. Run the script `mqft_setupCoordination.sh` on the coordination queue manager. For ex(update qmgr-pod-name as per your setup):  
	  `oc exec -ti <qmgr-pod-name> /etc/mqm/QMgrSetup/mqft_setupCoordination.sh`  
	  If the script is run successfully, then you will see an output containing set of MQSC commands run successfully.  
	  
####  3. Setting up Queue manager as Agent Queue manager  
 1. `mqft_setupAgent.sh` helps you automate the creation of all mft objects required on a queue manager for it to serve as  agent queue manager. It is recommended to run this script before we deploy the MFT Agent. This script takes in *agent name* as command line argument.   
 2. Run the script `mqft_setupAgent.sh` on the coordination queue manager. For ex(update qmgr-pod-name as per your setup):  
	```
	oc exec -ti <qmgr-pod-name> /etc/mqm/QMgrSetup/mqft_setupAgent.sh A1
	oc exec -ti <qmgr-pod-name> /etc/mqm/QMgrSetup/mqft_setupAgent.sh A2
	```   
#### 4. Build MFT Agent Docker image  
1. Perform all the below steps on oc-terminal  
2. Navigate to `<path>/mft-cp4i/AgentSetup`
3. Dowload MFT Agents redistributable package from [here] (https://www.ibm.com/support/pages/downloading-ibm-mq-version-915-continuous-delivery)  
	a. Scroll down the page to find **Clients** section and within that the **IBM MQ redistributable Managed File Transfer Agents** download link.  
	b.This redistributable package is used to build the MFT Agents  image to containerize MFT Agents.  
	c. Basic design is to have an MFT agent per container. This is done for ease of maintainability of MFT Agents.  
4. Run the command `docker build -t mftagent:1.0 -f Dockerfile-agent .`  
5. Once the docker image building is successful, move to next section  

#### 5. Push the MFT Agent Docker image to your cluster namespace
1. On oc-terminal run following commands  
2.  Run following commands to push the mft docker image to your cluster namespace  
	i. Please note to run `$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')'` and use the output of it as "route-name"  
	```
	docker tag mftagent:1.0  route-name/<cluster-namespace>/mftagent:1.0
	docker push route-name/<cluster-namespace>/mftagent:1.0
	```  

### 6. Deploy MFT Agent - A1
1. On oc-terminal, navigate to `<path>/mft-cp4i`  
2. Open the file mft_agentredit_Deployment.yaml and update following  
	a. metadata.name : This is the deployment name. This demo sets this as `mft-cp4i-a1`   
	b. Image repository: this is repository into which the mftagent docker image was pushed in step-5  
	c. Host, Port, Channel: Shall match to your qmgr setup  
	d. Agent name: This demo sets this as "A1"  
3. Run following command to deploy the mft agent  
	 `oc create -f mft_agentredit_Deployment.yaml`  
4. Run `oc get pods` to know if a deployment is created. If created, then you will a pod by name
`mft-cp4i-a1-xxxx-xxxx`   
	`ex: mft-cp4i-a1-6f7b555d4c-xdw9r`  

### 7. Deploy MFT Agent - A2  
1. Same as Step-6 above, except that update  
	a. metadata.name : Current deployment name. This demo sets this as `mft-cp4i-a2`   
	b. Agent name: This demo sets this as "A2"  
2. 	 Run `oc get pods` to know if a deployment is created. If created, then you will a pod by name
`mft-cp4i-a1-xxxx-xxxx`   
	`ex: mft-cp4i-a2-7ff4646fb6-z98z4`  

### 8. List Agents  

1. On the oc-terminal, run following  
2. On any one of the pods (mft-cp4i-a1-* or mft-cp4i-a2-*) run following command to list the current   agents. This tutorial uses A1 pod. For example:  
`oc exec -ti mft-cp4i-a1-6f7b555d4c-xdw9r /var/mqm/mft/bin/fteListAgents`  
3. On successful run, following output will be displayed  
```
[root@mqopr mft-cp4i]# oc exec -ti mft-cp4i-a1-6f7b555d4c-xdw9r /var/mqm/mft/bin/fteListAgents
5724-H72 Copyright IBM Corp.  2008, 2018.  ALL RIGHTS RESERVED

Command executed at 2020-06-04 10:34:30 UTC

Coordination queue manager time 2020-06-04 10:34:30 UTC

Agent Name:     Queue Manager Name:     Status:     Status Age:
A1              QUICKSTART              READY           0:01:42
A2              QUICKSTART              READY           0:02:26
```  

### 9.  Initial a file transfer from A1⇒ A2  
1. Execute all below step on the oc-terminal  
2. We will now, run a transfer from A1 Agent (pod - `mft-cp4i-a1-6f7b555d4c-xdw9r`) to A2 Agent (pod - `mft-cp4i-a2-7ff4646fb6-z98z4`). This means, a file is attempted to transfer from one pod to another pod  
3. As part of the setup, we created a sample file `/tmp/demofiles/samplefile.txt` in A1 pod only.   Inspect `<path>/AgentSetup/mqft.sh` file on how this was achieved  
4. We will now transfer this file `/tmp/demofiles/samplefile.txt` that's on A1 agent to A2 agent. To achieve this, run following command  
	`oc exec -ti mft-cp4i-a1-6f7b555d4c-xdw9r -- bash -c "/var/mqm/mft/bin/fteCreateTransfer -p QUICKSTART -sa A1 -sm QUICKSTART -da A2 -dm QUICKSTART -df /tmp/transfer.txt /tmp/demofiles/samplefile.txt"`  
5. On successful command run you will see following output  
```
[root@mqopr]# oc exec -ti mft-cp4i-a1-6f7b555d4c-xdw9r -- bash -c "/var/mqm/mft/bin/fteCreateTransfer -p QUICKSTART -sa A1 -sm QUICKSTART -da A2 -dm QUICKSTART -df /tmp/transfer.txt /tmp/demofiles/samplefile.txt"
5724-H72 Copyright IBM Corp.  2008, 2018.  ALL RIGHTS RESERVED
BFGCL0035I: Transfer request issued.  The request ID is: 414d5120515549434b53544152542020b68cd75e7b834324
BFGCL0182I: The request is now waiting to be processed by the agent.
[root@mqopr]#
```  

