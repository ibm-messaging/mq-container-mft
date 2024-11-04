#!/bin/bash

# © Copyright IBM Corporation 2024
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set +ex

# This is a simple shell script to setup MFT agents - one standard and a
# bridge agent. The commands in the shell script must be run manually in
# the sequence there given below.

# The script uses a bunch of json files and a shell script. All files must
# exist in the same directory as this script file.
#
# The value for "qmgrHost" key in sourceagentconfig.json and bridgeagentconfig.json 
# file must be set to the host name or IP address of the machine where your 
# queue manager is running. This script deploys an instance of the queue manager
# container on the same machine as agents.
#
# The values in the bridgeagentcrede.json file must be updated to point to your
# sftp server. The values for "serverHostKey" and "serverPrivateKey" keys must be
# base64 encoded, otherwise bridge agent will fail to connect to SFTP server.


# Get MQ images first.
podman pull icr.io/ibm-messaging/mq:latest

# Get MFT container image
podman pull icr.io/ibm-messaging/mqmft:latest

# Stop if queue manager container is already running
podman stop mftqm
# Delete container
podman rm mftqm

# Stop if MFT srcagent container is running
podman stop srcagent
# Delete container
podman rm srcagent
# Stop if MFT srcagent container is running
podman stop bridgeagent
# Delete container
podman rm bridgeagent

# Delete if secrets already exist
podman secret rm mqAdminPassword
podman secret rm mqAppPassword

# Create podman secret for MQ Container.
printf "passw0rd" | podman secret create mqAdminPassword -
printf "passw0rd" | podman secret create mqAppPassword -

# Start the MQ Container with container name mftqm
podman run --secret mqAdminPassword,type=mount,mode=0777 --secret mqAppPassword,type=mount,mode=0777 --env LICENSE=accept --env MQ_QMGR_NAME=MFTQM --publish 1414:1414 --publish 9443:9443 --detach --name mftqm icr.io/ibm-messaging/mq:latest

# Copy queue manager configuration file for MFT and execute it to create MFT configuration. 
# You may have do chmod 777 ./qmconfig.mqsc so that other users are able to execute inside the container
podman cp ./qmconfig.mqsc mftqm:/run
# Copy a shell script containing commands to setup authorities on queue manager objects
# You may have do chmod 777 ./setauth.sh so that other users are able to execute inside the container
podman cp ./setauth.sh mftqm:/run

# Create objects required for MFT.
podman exec -it mftqm /bin/bash -c "runmqsc MFTQM < /run/qmconfig.mqsc"
# Set authorities on qm objects so that agent running MFT Container can connect to queue manager
podman exec -it mftqm /bin/bash -c "/run/setauth.sh"

# Run a standard agent. The agent configuration JSON file in the current directory is mounted into the container.
podman run --volume "${PWD}":/mftagentcfg/agentcfg --env BFG_JVM_PROPERTIES="-Djava.util.prefs.systemRoot=/jprefs/.java/.systemPrefs -Djava.util.prefs.userRoot=/jprefs/.java/.userPrefs" --env LICENSE=accept --env MFT_AGENT_NAME=SRC --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/sourceagentconfig.json --detach --name srcagent icr.io/ibm-messaging/mqmft:latest

# Run a bridge agent. The agent configuration JSON file and credential files in the current directory is mounted into the container.
podman run --volume "${PWD}":/mftagentcfg/agentcfg --volume "${PWD}":/mnt/credentials --env BFG_JVM_PROPERTIES="-Djava.util.prefs.systemRoot=/jprefs/.java/.systemPrefs -Djava.util.prefs.userRoot=/jprefs/.java/.userPrefs" --env LICENSE=accept --env MFT_AGENT_NAME=BRIDGE --env MFT_AGENT_CONFIG_FILE=/mftagentcfg/agentcfg/bridgeagentconfig.json --detach --name bridgeagent icr.io/ibm-messaging/mqmft:latest

