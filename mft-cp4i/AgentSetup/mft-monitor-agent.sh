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

state()
{
  ftePingAgent -m ${MQ_QMGR_NAME} -w 10 ${MFT_AGENT_NAME} | awk -F ':' '/BFGCL0213I/{print $1}; NR in nr'
}
trap "source mft-stop-container.sh" SIGTERM SIGINT
echo "Monitoring MFT Agent" ${MFT_AGENT_NAME}

# Loop until "ftePingAgent" says the MFT Agent is running
until [ "`state`" == "BFGCL0213I" ]; do
  sleep 1
done
ftePingAgent -m ${MQ_QMGR_NAME} -w 10 ${MFT_AGENT_NAME}
echo "IBM MFT Agent ${MFT_AGENT_NAME} is now fully running"

until [ "`state`" != "BFGCL0213I" ]; do
  sleep 5
done

fteListAgents -p ${MQ_QMGR_NAME}
