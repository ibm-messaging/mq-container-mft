---
copyright:
  years: 2017, 2020
lastupdated: "2020-07-28"
---

# MQ Managed File Transfer for Containers

#### What is IBM MQ Managed File Transfer(MFT) ?
For many organizations, the exchange of files between business systems remains a common and important integration methodology. Files are the simplest unit of data to exchange and often represent the lowest common denominator for an enterprise infrastructure.

Although the exchange of files is conceptually simple, doing so in the enterprise is a challenge to manage and audit. This difficulty is brought into clear focus when an organization needs to perform file transfer with a different business organization, perhaps using a different physical network, with different security requirements, and perhaps a different governance or regulatory framework.

IBM® MQ File Transfer Edition provides an enterprise-grade managed file transfer capability that is both robust and easy to use. MQ File Transfer Edition exploits the proven reliability and connectivity of MQ to transfer files across a wide range of platforms and networks. MQ File Transfer Edition takes advantage of existing MQ networks, and you can integrate it easily with existing file transfer systems.

You can find more information at [IBM Knowledge Centre](https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_9.2.0/com.ibm.mq.pro.doc/wmqfte_intro.htm)

---

## Running MFT Agents in Containers
MQ Advanced supports running MFT Agents in docker containers and this guide will help you setup MFT for docker, run MFT agents in containers and also run a successful file transfer using  MFT agents running in containers.

## Prerequisites

* [Install Docker](https://docs.docker.com/install/linux/docker-ce/ubuntu/)
* Clone the repository **mft-containers** into your computer:
    1. Open a new command-shell and navigate to home path
        ```
        On Linux: cd /home
        On Windows: cd %HOMEDRIVE%
        ```
    2. Clone the current repository into **mft-containers** directory
        ```
        HTTPS: git clone https://github.com/ibm-messaging/mft-cloud.git
        SSH: git clone git@github.com:ibm-messaging/mft-cloud.git
        ```
* Download **9.2.0.0-IBM-MQFA-Redist-LinuxX64.tar.gz** from [IBM Fixcentral](https://www.ibm.com/support/fixcentral/) into mft-containers/agent/
**Note: 9.2.0.0-IBM-MQFA-Redist-LinuxX64.tar.gz** has to be in same path as **Dockerfile**.

---
### MFT Container Customisations  

We will add few customizations that would simplify the configuration of MFT for coordination queue manager and agent setup.
**Note:** These custimization are only needed to simplify the MFT configuration. The same steps, that are set of mqsc commands, can be executed on base mq image directly using *docker exec*.  

#### Understanding Customizations  
1. Configuring MFT Coordination Queue manager:

    1.1 Coordination manager requires set of system queues and topics to be created, this is one time activity.  
    1.2 Agents will need a SVRCONN channel using which they can communication with MQ queue manager. This SVRCONN channel has to be setup with appropriate CHLAUTH/CONNAUTH rules. This guide creates a new channel **MFT_SVRCONN** and setup CHLAUTH/CONNAUTH rules such that any user with a valid password can connect to queue manager.  
    1.3 **mft_setupCoordination.sh** available in **mft-containers/server** directory is aimed at simplifying this configuration.  
    * *mft-containers/server/setupQMForMFTAgents.mqsc* : MQSC script file that has definations for new resources.  
    
2. New Agent Setup:  
Every MFT Agent needs set of system queues to be created on the coordination queue manager.  
    2.1 **mft_setupAgent.sh** available in **mft-containers/server/** directory is aimed at simplifying this configuration
     - *mft-containers/server/createAgent.mqsc*: MQSC file that has definitions for creating new queues for Agent.  

    2.2 **mft-containers/server/mft_removeAgent.sh** available in **server** directory is to remote the agent resuorces from the coordination manager. This is to be executed if an mft agent has to be deleted.
    - *deleteAgent.mqsc*: MQSC script file that has definitions for delete Agent's resouces on coordination queue manager.  

All the cutomizations (.sh and .mqsc) files are copied into **/etc/mqm/mft/** path of the container.

---

### Creating Customized MQ Queue Manager for MFT in containers

1. Open a command shell and navigate to path of **mft-containers/server**. For example */home/mft-containers/server*(on Linux) or *C:\\mft-containers\\server*(on Windows).  
2. Run a docker command to make sure docker is setup and docker service is running
    ```
    docker version
    docker ps -a
    ```
    First command will print docker information and second command will list all the available containers in your docker environment.
3. Create a new customized mq image
    ```
    docker build -t mqadvmft  -f Dockerfile-server
    ```
4. Once the docker build is successful, run a new container of it, which is queue manager in container. This queue manager is to be used as coordination queue manager.
    ```
    docker run --env LICENSE=accept --env MQ_QMGR_NAME=QM1 --publish 1414:1414 --publish 9443:9443 --detach --name=QM1 mqadvmft
    ```
5. Run the below command to list the docker containers, find newly created container **QM1** and make a note of its container-id.
    ```
    docker ps
    ```
6. Setup the **QM1** as coordination queue manager. Running the **mft_setupCoordination.sh** script will create required configuration.
    ```
    docker exec -ti <QM1-container-id> /etc/mqm/mft/mqft_setupCoordination.sh
    ```
    Successful run of this script completes queue manager configuration for MFT.  
    
---

### Agent setup for MFT Containers

Agent package contains a Dockerfile-agent to build the MFT agent docker image. MFT Agent is setup and started as part of the **mqft.sh** script. This package assumes a single queue manager(**QM1**) as coordination queue manager, command queue manager and agent queue manager.
**Note:** As per your application architecture, you can consider to have separate queue managers for coordination, command and agents.

1. Copy the MFT redistributable package: **9.2.0.0-IBM-MQFA-Redist-LinuxX64.tar.gz** to the Agent directory. For example 
    ```
    On Linux: /home/mft-containers/agent 
    On Windows: %HOMEDRIVE$\mft-containers/agent 
    ```
2. Open a command shell and navigate to path of **mft-containers/agent** repository. For example */home/mft-containers/server*(on Linux) or *C:\\mft-containers\\server*(on Windows).  

3. Build IBM MQ Managed File Transfer agent docker image.
   This MQ MFT docker image uses a golang application to create an agent and monitor status of agent. The agent configuration is specified using a JSON file. The JSON file must be located on a persistent volume. 
    ```
    docker build -t mftagentredist -f Dockerfile-agent .
    ```
4. We will create two agents as part of this document to demonstrate mft agents in container. These agents are **AGENTSRC** and **AGENTDEST**. As a first step of agent configuration, we have to create their congfiguration on coordination queue manager.
    ```
    docker exec -ti <QM1-container-id> /etc/mqm/mft/mqft_setupAgent.sh AGENTSRC
    docker exec -ti <QM1-container-id> /etc/mqm/mft/mqft_setupAgent.sh AGENTDEST
    ```
    **Note:** 
    1. QM1-Container-id is the id of the queue manager container created in above section.
    2. **mqft_setupAgent.sh** script requires MFT agent name as input parameter
    3. To configure a IBM MQ Managed file transfer protocol Bridge Agent(PBA agent) [click here](./README_pbagent.md) for the steps.

5. We will create a volume on the host system. This volume will be mounted to container and used as persistent storage for agent configuration and logs. The JSON file for setting up an agent can be specified in this volume itself.
    ```
    docker volume create mftdata 
    ```
   Verify the volume creation by
    ```
    docker volume inspect mftdata 
    ```
   
6. Once the docker-agent build is successful, run a new container of it, which is agent in container. 
    ```
    docker run --mount type=volume,source=mftdata,target=/mftdata -e AGENT_CONFIG_FILE="/mftdata/agentconfigsrc.json" -d --name=AGENTSRC mftagentredist
    docker run --mount type=volume,source=mftdata,target=/mftdata -e AGENT_CONFIG_FILE="/mftdata/agentconfigdest.json" -d --name=AGENTDEST mftagentredist   
    ```
    **Note:** 
    1. AGENT_CONFIG_FILE is the environment variable that points to a JSON file containing required information for creating an agent. The path will be on persistent volume mounted on a container.
    2. mftagentredist: Is the docker image of mft redistributable agents.
    
	While most the attributes of JSON file are self explanatory, here is a brief explanation of on some of them.
    ```
	dataPath: Absolute path where agent configuration will be created.
	agentMonitorInterval: Frequency at which container will monitor the status of an agent. A monitor program 'mftcfg' is used to monitor the status of agent.
	displayAgentLogs: Read and display logs on console from output0.log file of an agent.
	maximumDisplayLines: The number of lines from the output0.log file to display. The latest lines from the log file will be displayed.
    ```
	
	The following is an example of an agent configuration JSON file
    ```
	{
      "dataPath" : "/mftdata",
      "agentMonitorInterval" : 150,
      "displayAgentLogs" : false,
      "maximumDisplayLines" : 50,
      "coordinationQMgr" : {
        "name":"MFTQM",
        "host":"coordqmhost.com",
        "port":1414,
        "channel":"MFT.CHN"
      },
      "commandsQMgr" : {
        "name":"MFTQM",
        "host":"cmdqmhost.com",
        "port":1414,
        "channel":"MFT.CHN"
      },
      "agent" : {
        "name":"AGNTSRC",
        "agentType" : "STANDARD",
        "qmgrName":"MFTQM",
        "qmgrHost":"agentqmhost.com",
        "qmgrPort":1414,
        "qmgrChannel":"MFT.CHN",
        "credentialsFile":"/usr/local/bin/MQMFTCredentials.xml",
	    "protocolBridge" : {
		  "credentialsFile":"/usr/local/bin/ProtocolBridgeCredentials.xml",
		  "serverType":"SFTP",
		  "serverHost":"mysftp.com",
		  "serverTimezone":"",
		  "serverPlatform":"UNIX",
		  "serverLocale":"en-US",
		  "serverFileEncoding":"UTF-8",
		  "serverPort":22,
		  "serverTrustStoreFile" : "",
		  "serverLimitedWrite":"",
		  "serverListFormat" :"",
		  "serverUserId":"sftpuser",
		  "serverPassword":"sftpPassw@rd"
	    }
	  }
   }
    ```

7. Run the below command to list the docker containers, find newly created containers **AGENTSRC**, **AGENTDEST** and make a note of their container-ids.
    ```
    docker ps
    ```
8. Check if the mft agents accept commands and show output. For example, run following command on Agent containers
    ```
    docker exec -ti <AGENTSRC-container-id> fteListAgents
    docker exec -ti <AGENTDEST-container-id> fteListAgents
    ```
    If both the containers show the Agents information, that confirms Agent's running in containers and accepting commands.
8. To verify the Agent setup, use following command to list the container logs
    ```
    docker logs <AGENTSRC-container-id>
    docker logs <AGENTDEST-container-id>
    ```
---

### Create a File Transfer with MFT Agents in Containers

We will create a text file on **AGENTSRC** that will be transferred to **AGENTDEST**. We then transfer this file using the **fteCreateTransfer** command.

1. Open a command shell and run the docker command to list all the containers. Check that newly created containers(**QM1,AGENTSRC and AGENTDEST**) are running.
    ```
    docker ps
    ```
2. Create a text file on **AGENTSRC** for file transfer.
    ```
    docker exec -ti <AGENTSRC-container-id> bash -c "echo 'Hello World' >> /tmp/file.txt"
    ```
3. Check if the file is created on **AGENTSRC**
    ```
    docker exec -ti <AGENTSRC-container-id> bash -c "cat /tmp/file.txt"
    ```
4. Check that file doesn't exist on **AGENTDEST**.
    ```
    docker exec -ti <AGENTDEST-container-id> bash -c "cat /tmp/transfer.txt"
    ```
    **Note:** Since the transfer.txt file doesn't exist on **AGENTDEST**, above command may result in an error. This error can be ignored for now.
5. Run the file transfer command to transfer newly created file (**file.txt**) to **AGENTDEST**.
    ```
    docker exec -ti <AGENTSRC-container-id> fteCreateTransfer -p QM1 -sa AGENTSRC -sm QM1 -da AGENTDEST -dm QM1 -df /tmp/transfer.txt /tmp/file.txt
    ```
6. Wait for couple of seconds and check if file is transferred to **AGENTDEST**.
    ```
    docker exec -ti <AGENTDEST-container-id> bash -c "cat /tmp/transfer.txt"
    ```
    **Note:** If the output of above command is *Hello World*, that confirms file transfer is complete and successful.

---
### Conclusion  

As part of this document we have created a customized MQ image for MFT, started MFT agents in containers and demonstrated that file transfer runs between the agents in containers and transferred file exists in destination agent.