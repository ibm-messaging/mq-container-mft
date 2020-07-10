#!/bin/bash
# -*- mode: sh -*-
# © Copyright IBM Corporation 2015, 2020
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

set -e
su samantha

cd /var/mqm/mft/bin

export PATH=$PATH:/var/mqm/mft/bin:/var/mqm/mft/java/jre64/jre/bin
export MQ_QMGR_NAME=${ENV_AGQMGR}
export MQ_QMGR_HOST=${ENV_AGQMGRHOST}
export MQ_QMGR_PORT=${ENV_AGQMGRPORT}
export MQ_QMGR_CHL=${ENV_AGQMGRCHN}
export MFT_AGENT_NAME=${ENV_AGNAME}
export BFG_DATA=${ENV_BFGDATA}

echo "Setting up FTE Environment for this Agent : " ${BFG_DATA}
cp -f /usr/local/bin/MQMFTCredentials.xml  $HOME
cp -f /usr/local/bin/ProtocolBridgeCredentials.xml $HOME
chmod go-rw $HOME/MQMFTCredentials.xml
chmod go-rw $HOME/ProtocolBridgeCredentials.xml
chmod go-rwx /usr/local/bin/MQMFTCredentials.xml
pwd
ls -l $HOME

#source fteCreateEnvironment

# Assumption is that a single queue manager will be used as Coorindation/Command/Agent.
# User should have passed us the queue manager details as environment variables.
echo "Setting up Coordination manager for this agent"
fteSetupCoordination -coordinationQMgr ${MQ_QMGR_NAME} -coordinationQMgrHost ${MQ_QMGR_HOST} -coordinationQMgrPort ${MQ_QMGR_PORT} -coordinationQMgrChannel ${MQ_QMGR_CHL} -f
echo "Coordination manager setup completed"

echo "Setting up Command manager for this agent"
fteSetupCommands -connectionQMgr  ${MQ_QMGR_NAME} -connectionQMgrHost ${MQ_QMGR_HOST} -connectionQMgrPort  ${MQ_QMGR_PORT} -connectionQMgrChannel ${MQ_QMGR_CHL} -f
echo "Command manager setup completed"

if ["${IS_PBA_AGENT}" == "true"]; then
  echo "Creating a MFT PBA Agent"
  fteCreateBridgeAgent -agentName ${MFT_AGENT_NAME} -agentQMgr ${MQ_QMGR_NAME} -agentQMgrHost ${MQ_QMGR_HOST} -agentQMgrPort ${MQ_QMGR_PORT} -agentQMgrChannel ${MQ_QMGR_CHL} -bt ${PROTOCOL_FILE_SERVER_TYPE} -bh ${SERVER_HOST_NAME} -btz ${SERVER_TIME_ZONE} -bm ${SERVER_PLATFORM} -bsl ${SERVER_LOCALE} -bfe ${SERVER_FILE_ENCODING} -credentialsFile "/usr/local/bin/MQMFTCredentials.xml" -f
  echo "MFT PBA Agent creation was successful"
else
  echo "Creating a MFT Agent"
  fteCreateAgent -agentName ${MFT_AGENT_NAME} -agentQMgr ${MQ_QMGR_NAME} -agentQMgrHost ${MQ_QMGR_HOST} -agentQMgrPort ${MQ_QMGR_PORT} -agentQMgrChannel ${MQ_QMGR_CHL} -credentialsFile "/usr/local/bin/MQMFTCredentials.xml" -f
  echo "Agent creation was successful"
fi


if [ "${MFT_AGENT_NAME}" == "A1" ]; then
  echo "Creating a sample file"
  mkdir -p /tmp/demofiles
  echo "This is a demo file created to test mft file transfer" > /tmp/demofiles/samplefile.txt
else 
  mkdir -p /tmp/dropboxfiles
fi

# Enable queue input out from this agent
echo "enableQueueInputOutput=true" >> $BFG_DATA/mqft/config/$MQ_QMGR_NAME/agents/$MFT_AGENT_NAME/agent.properties
echo "highlyAvailable=true" >> $BFG_DATA/mqft/config/$MQ_QMGR_NAME/agents/$MFT_AGENT_NAME/agent.properties

#echo "Starting MFT Agent...."
fteStartAgent -p  ${MQ_QMGR_NAME} ${MFT_AGENT_NAME}
echo "MFT Agent Started"

fteListAgents -p ${MQ_QMGR_NAME}

fteCreateTransfer -gt task.xml -sa KXAGNT -sm MFTHAQM -da KXAGNT -dm MFTHAQM -dq "SWIFTQ@MFTHAQM" -qs 1K "/mftdata/xferdata/source.txt"
fteCreateMonitor -ma KXAGNT -mn FILEMON -md "/mftdata/trigger" -tr "match,*.txt" -f -mt task.xml


# Monitor a particular directory to upload files to dropbox.
cd /mftdkr
echo "Starting monitoring application"
# Run Java monitoring application

exec java -cp . monitor.MftAgent ${MFT_AGENT_NAME}

