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

export GOPATH=/usr/local/go
export PATH=$PATH:/var/mqm/mft/bin:/var/mqm/mft/java/jre64/jre/bin:$GOPATH/bin
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

