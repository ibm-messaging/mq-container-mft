#!/bin/bash
# -*- mode: sh -*-
# � Copyright IBM Corporation 2015, 2017
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

extnCreate=_create.mqsc
extnDelete=_delete.mqsc
MQ_MFT_AGENT=$1
echo "Setting up Agent:"$MQ_MFT_AGENT
#copy create file
cp /etc/mqm/QMgrSetup/createAgent.mqsc /etc/mqm/QMgrSetup/$MQ_MFT_AGENT$extnCreate
sed -i -e "s/agentName/$MQ_MFT_AGENT/g" /etc/mqm/QMgrSetup/$MQ_MFT_AGENT$extnCreate

#copy delete file
cp /etc/mqm/QMgrSetup/deleteAgent.mqsc /etc/mqm/QMgrSetup/$MQ_MFT_AGENT$extnDelete
sed -i -e "s/agentName/$MQ_MFT_AGENT/g" /etc/mqm/QMgrSetup/$MQ_MFT_AGENT$extnDelete

#run only create.
runmqsc ${MQ_QMGR_NAME} < /etc/mqm/QMgrSetup/$MQ_MFT_AGENT$extnCreate
echo "Completed..."
