#!/bin/bash
# -*- mode: sh -*-
# © Copyright IBM Corporation 2015, 2017
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
cd /var/mqm/mft/bin

export PATH=$PATH:/var/mqm/mft/bin

echo "Setting up FTE Environment for this Agent : " ${BFG_DATA}
source fteCreateEnvironment

echo "Agent Configuration Details:"
echo "Home directory list:"
ls -l $HOME

echo "Queue manager:" $MQ_QMGR_NAME
echo "Queue manager host:" $MQ_QMGR_HOST
echo "Queue manager port:" $MQ_QMGR_PORT
echo "Queue manager channel:" $MQ_QMGR_CHL

# Assumption is that a single queue manager will be used as Coorindation/Command/Agent.
# User should have passed us the queue manager details as environment variables. 

echo "Setting up Coordination manager for this agent"
fteSetupCoordination -coordinationQMgr ${MQ_QMGR_NAME} -coordinationQMgrHost ${MQ_QMGR_HOST} -coordinationQMgrPort ${MQ_QMGR_PORT} -coordinationQMgrChannel ${MQ_QMGR_CHL} -f
echo "Coordination manager setup completed"

echo "Setting up Command manager for this agent"
fteSetupCommands -connectionQMgr  ${MQ_QMGR_NAME} -connectionQMgrHost ${MQ_QMGR_HOST} -connectionQMgrPort  ${MQ_QMGR_PORT} -connectionQMgrChannel ${MQ_QMGR_CHL} -f
echo "Command manager setup completed"

echo "Creating MFT Agent"
fteCreateAgent -agentName ${MFT_AGENT_NAME} -agentQMgr ${MQ_QMGR_NAME} -agentQMgrHost ${MQ_QMGR_HOST} -agentQMgrPort ${MQ_QMGR_PORT} -agentQMgrChannel ${MQ_QMGR_CHL} -credentialsFile "$HOME/MQMFTCredentials.xml" -f
echo "Agent creation was successful"


if [ "${MFT_AGENT_NAME}" == "A1" ]; then
  echo "Creating a sample file"
  mkdir -p /tmp/demofiles
  echo "This is a demo file created to test mft file transfer" > /tmp/demofiles/samplefile.txt
fi

#echo "Starting MFT Agent...."
fteStartAgent -p  ${MQ_QMGR_NAME} ${MFT_AGENT_NAME}
echo "MFT Agent Started"

fteListAgents -p ${MQ_QMGR_NAME}

#Wait before monitoring agent for container
sleep ${MFT_WAIT_FOR_AGENT}

# Monitor a particular directory to upload files to dropbox.
mft-monitor-agent.sh
