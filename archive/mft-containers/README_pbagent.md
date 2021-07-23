---
copyright:
  years: 2017, 2020
lastupdated: "2020-05-19"
---

# MQ Managed File Transfer for Containers

### Protocol Bridge Agent setup for MFT Containers

4. We will create pba agent as part of this document to demonstrate mft agents in container. This agents is named **AGENTPBA**. As a first step of agent configuration, we have to create their congfiguration on coordination queue manager.
    ```
    docker exec -ti <QM1-container-id> /etc/mqm/mft/mqft_setupAgent.sh AGENTPBA
    ```
    **Note:** 
    1. QM1-Container-id is the id of the queue manager container created in above section.
    2. **mqft_setupAgent.sh** script requires MFT agent name as input parameter.

5. Once the docker-agent build is successful, run a new container of it, which is agent in container. 
    ```
    docker run --env MQ_QMGR_NAME=QM1  --env MQ_QMGR_HOST=<docker-host-ip> --env MQ_QMGR_PORT=1414 --env MQ_QMGR_CHL=MFT.SVRCONN --env MFT_AGENT_NAME=AGENTPBA --env IS_PBA_AGENT=true --env PROTOCOL_FILE_SERVER_TYPE=<userinputvalues> --env SERVER_HOST_NAME=<userinputvalues> --env SERVER_TIME_ZONE=<userinputvalues> --env SERVER_PLATFORM=<userinputvalues> --env SERVER_LOCALE=<userinputvalues> --env SERVER_FILE_ENCODING=<userinputvalues> -d --name=AGENTPBA mftagentredist
    ```
    - Above command creates a template file **ProtocolBridgeProperties.xml**, in the agent configuration directory MQ_DATA_PATH/mqft/config/coordination_queue_manager/agents/bridge_agent_name. The command also creates an entry in the file for the default protocol file server, if a default was specified when the command was run.
    - To Specify more than one server configuration in the **ProtocolBridgeProperties.xml** file you will have to [`docker exec`](https://docs.docker.com/engine/reference/commandline/exec/) into the pba agent container and modify the file and then save a copy the updated docker container for the further use. To do this you can use the `docker commit` command which saves the container state along with all its changes. for more information [click here](https://docs.docker.com/engine/reference/commandline/commit/) .

    **Note:** 
    1. <docker-host-ip>: Is the IP Address of the docker host. This could be found out by running `docker inspect` command and look for ipv4address field.
    2. MQ_QMGR_NAME=QM1: Is the queue manager we created and configured as coordination queue manager in above section.
    3. mftagentredist: Is the docker image of mft redistributable agents.
    4. the \"<userinputvalues\>" are the configuration values required for setting up the PBA Agent. For more detailed information [click here](https://www.ibm.com/support/knowledgecenter/SSFKSJ_9.1.0/com.ibm.mq.ref.adm.doc/create_bridge_agent_cmd.htm) .
    5. modify the **ProtocolBridgeCredentials.xml** file with your server credentials before building the docker image.