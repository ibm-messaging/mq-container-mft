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

# This script deploys MFT Agents on a OpenShift Cluster.

# This script expects:
#   1) oc client is installed on the machine where this script will be executed.
#   2) User has done a oc login with valid credentials.
#   3) MQ Operator 3.2 or latest to be installed ibm-mqmft namespace.
#   4) Namespace ibm-mqmft is created in OpenShift cluster.
#

#Retry count
maxRetries=$1
#Retrt delay
retryDelay=$2
# By default images are pulled from icr.io. Override the image url if required.
imageURL=$3

# Source agent name
SRC_AGENT_NAME=SRC
# Bridge agent name
BRIDGE_AGENT_NAME=BRIDGE
# Destination agent name
DEST_AGENT_NAME=DEST
# Agent, command and coordination queue manager
QM_NAME=MFTQM

# OCP Namespace
OC_NAMESPACE=ibm-mqmft
# Queue Manager deployment Name
OC_QM_NAME=mftqm
# Bridge Agent deployment name
OC_BRIDGE_AGENT="bridge-agent"
# Source Agent deployment name
OC_SRC_AGENT="source-agent"
# Destination agent deployment name
OC_DEST_AGENT="dest-agent"
# Persistent volume claim for source agent
OC_PVC_SOURCE="pvc_source"
# Persistent volume claim for destination agent
OC_PVC_DEST="pvc_dest"

# Queue manager configuration
OC_QM_CONFIG_MAP=./qmconfig.yaml
# Queue manager deployment
OC_QM_DEPLOYMENT=./qm.yaml
# Source agent configuration attributes
OC_SOURCE_AGENT_CONFIG_MAP=./source-agent-config.yaml
# Source agent configMap name
OC_SOURCE_AGENT_CONFIG_MAP_NAME="src-agent-config"
# Bridge agent configMap name
OC_BRIDGE_AGENT_CONFIG_MAP_NAME="pba-config"
# Bridge agent credentials configMap name
OC_BRIDGE_AGENT_CRED_CONFIG_MAP_NAME="pba-custom-cred"
# Destination agent configMap name
OC_DEST_AGENT_CONFIG_MAP_NAME="dest-agent-config"

# Bridge agent configuration attributes
OC_BRIDGE_AGENT_CONFIG_MAP=./bridge-agent-config.yaml
# Destination agent configuration attributes
OC_DEST_AGENT_CONFIG_MAP=./dest-agent-config.yaml
# Credentials for bridge agent
OC_BRIDGE_AGENT_CRED_CONFIG_MAP=./pba-cust-cred-map.yaml
# Source agent deployment
OC_SRC_AGENT_DEPLOYMENT=./source-agent-deployment.yaml
# Bridge agent deployment
OC_BRIDGE_AGENT_DEPLOYMENT=./bridge-agent-deployment.yaml
# Destination agent deployment
OC_DEST_AGENT_DEPLOYMENT=./dest-agent-deployment.yaml
# Source agent PVC
OC_SOURCE_PVC="./pvc-source.yaml"
# Destination agent PVC
OC_DEST_PVC="./pvc-dest.yaml"

# Validation of input parameters
if [ -z "$maxRetries" ] ||  [ -z "$retryDelay" ] || [ "$maxRetries" -lt 0 ] || [ "$retryDelay" -lt 0 ]; then
    echo "Usage ./ocp-deploy <Max retries> <Delay between each retry> <imageURL=<url>>"
    exit 1
fi

if [ -n "$imageURL" ]  ; then
    echo "Deploying image from $imageURL"

    # Make a copy of deployment yamls and update image url
    cp -f source-agent-deployment.yaml ../run
    cp -f bridge-agent-deployment.yaml ../run
    # shellcheck disable=SC2086
    sed -i 's|'imageURL-url'|'${imageURL}'|g' ../run/source-agent-deployment.yaml
    # shellcheck disable=SC2086
    sed -i 's|'imageURL-url'|'${imageURL}'|g' ../run/bridge-agent-deployment.yaml
    OC_SRC_AGENT_DEPLOYMENT="../run/source-agent-deployment.yaml"
    OC_BRIDGE_AGENT_DEPLOYMENT="../run/bridge-agent-deployment.yaml"
fi

# STEP 1: Switch to ibm-mqmft namespace
ocrc=$(oc project "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -gt 0 ] ; then
    echo "ibm-mqmft namespace may not exist. Command failed with return code $?. Script will end now. $ocrc"
    exit 1
fi

# Step 2: Create configMap containing MQSC commands required to created agent queues and topics
echo "Creating queue manager configMap $OC_QM_CONFIG_MAP"
ocrc=$(oc apply -f $OC_QM_CONFIG_MAP > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -gt 0 ] ; then
    echo "Failed to create queue manager configuration configMap $OC_QM_CONFIG_MAP. Command failed with return code $?. Script will end now."
    exit 1
fi

# Step 3: Deploy queue manager.
echo "Checking if queue manager $QM_NAME already running in namespace $OC_NAMESPACE"
ocrc=$(oc get queuemanager $OC_QM_NAME -n $OC_NAMESPACE > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    echo "Queue manager $QM_NAME already exists."
else
    grep "(NotFound)" <<< "$ocrc"
    # shellcheck disable=SC2181
    if [ "$?" -gt 0 ] ; then
        echo "Queue manager $QM_NAME does not exist. Creating one now."
        ocrc=$(oc apply -f $OC_QM_DEPLOYMENT -n "${OC_NAMESPACE}" > /dev/null 2>&1)
        # shellcheck disable=SC2181
        if [ $? -gt 0 ] ; then
            echo "Failed to deploy queue manager $QM_NAME. Command failed with return code $?. Script will end now."
            exit 1
        fi
    fi
fi


# Step 4: Verify if the QM is up and running
echo "Verifying status of queue manager $QM_NAME"
for (( i=1; i <= maxRetries; ++i ))
do
    qmstatus=$(oc get queuemanager "$OC_QM_NAME" -n "${OC_NAMESPACE}" )
    if grep -q "Running" <<< "$qmstatus"; then
        echo "The queue manager ${QM_NAME} is Running"
        break
    else
        if [ $i = maxRetries ] ; then
            echo "The queue manager ${QM_NAME} failed to start. Analyze the queue manager logs and fix the issue(s). Script will now end. $qmstatus"
            exit 1
        else
            echo "The queue manager ${QM_NAME} is not yet running. Will check the status again in $retryDelay seconds."
            sleep "$retryDelay"
        fi
    fi
done

# Step 5: Create PVC for source agent
echo "Creating Persistent Volume Claim $OC_PVC_SOURCE for agent $SRC_AGENT_NAME"
ocrc=$(oc apply -f $OC_SOURCE_PVC -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    echo "Failed to persistent volume for source agent. Command failed with return code $?. Script will end now. $ocrc"
    exit 1
else
    echo "$ocrc"
fi

# Step 6: Create PVC for destination agent
echo "Creating Persistent Volume Claim $OC_PVC_DEST for agent $DEST_AGENT_NAME"
ocrc=$(oc apply -f $OC_DEST_PVC -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    echo "Failed to persistent volume for source agent. Command failed with return code $?. Script will end now. $ocrc"
    exit 1
else
    echo "$ocrc"
fi

# Step 7: Create configMap cotaining attributes required to create a standard agent in JSON file
echo "Creating configMap $OC_SOURCE_AGENT_CONFIG_MAP_NAME for agent $SRC_AGENT_NAME"
ocrc=$(oc apply -f $OC_SOURCE_AGENT_CONFIG_MAP -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    echo "Failed to deploy standard agent configMap. Command failed with return code $?. Script will end now. $ocrc"
    exit 1
else
    echo "$ocrc"
fi

# Step 8: Create configMap cotaining attributes required to create a standard agent in JSON file
echo "Creating configMap $OC_DEST_AGENT_CONFIG_MAP_NAME for agent $DEST_AGENT_NAME"
ocrc=$(oc apply -f $OC_DEST_AGENT_CONFIG_MAP -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    echo "Failed to deploy standard agent configMap. Command failed with return code $?. Script will end now. $ocrc"
    exit 1
else
    echo "$ocrc"
fi

# Step 9: Create configMap cotaining attributes required to create a bridge agent in JSON file
echo "Creating configMap $OC_BRIDGE_AGENT_CONFIG_MAP_NAME for agent $BRIDGE_AGENT_NAME"
ocrc=$(oc apply -f $OC_BRIDGE_AGENT_CONFIG_MAP -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    echo "Failed to deploy bridge agent configMap. Command failed with return code $?. Script will end now. $ocrc"
    exit 1
else
    echo "$ocrc"
fi

# Step 10: Create configMap for bridge agent credentials
echo "Creating configMap $OC_BRIDGE_AGENT_CRED_CONFIG_MAP_NAME containing credentials for bridge agent $BRIDGE_AGENT_NAME"
ocrc=$(oc apply -f $OC_BRIDGE_AGENT_CRED_CONFIG_MAP -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    echo "Failed to deploy custom credential configMap for bridge agent. Command failed with return code $?. Script will end now. $ocrc"
    exit 1
else
    echo "$ocrc"
fi

# Step 11: Deploy standard agent
echo "Deploying agent $SRC_AGENT_NAME"
ocrc=$(oc apply -f $OC_SRC_AGENT_DEPLOYMENT -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    echo "Failed to deploy Source Agent $SRC_AGENT_NAME. Command failed with return code $?. Script will now end. $ocrc"
    exit 1
else
    echo "$ocrc"
fi

# Step 12: Verify the Source Agent pod is up and running
echo "Verifying status of agent $SRC_AGENT_NAME"
for (( i=1; i <= maxRetries; ++i ))
do
    agentStatus=$(oc get deployment ${OC_SRC_AGENT} -n "${OC_NAMESPACE}")
    # shellcheck disable=SC2181
    if [ "$?" -eq 0 ]; then
        # Check pod status. it must be 1/1
        if grep -q "1/1" <<< "$agentStatus"; then
            echo "The agent $SRC_AGENT_NAME is running"
            break
        else
            echo "The agent ${SRC_AGENT_NAME} is not yet running. Will check the status again in $retryDelay seconds."
            sleep "$retryDelay"
        fi
    else
        if [ $i = maxRetries ] ; then
            echo "The agent ${SRC_AGENT_NAME} failed to start within the timeout period. Analyze agent deployment or pod logs and fix the issue. Script will now end. $agentStatus" 
            exit 1
        fi
    fi
done

# Step 13: Deploy bridge agent
echo "Deploying agent $BRIDGE_AGENT_NAME"
ocrc=$(oc apply -f $OC_BRIDGE_AGENT_DEPLOYMENT -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    # shellcheck disable=SC2319
    echo "Failed to deploy bridge agent $BRIDGE_AGENT_NAME. Command failed with return code $?. Script will now end. $ocrc"
    exit 1
fi

# Step 14: Verify the bridge agent pod is up and running
echo "Verifying status of agent $BRIDGE_AGENT_NAME"
for (( i=1; i <= maxRetries; ++i ))
do
    agentStatus=$(oc get deployment "${OC_BRIDGE_AGENT}" -n "${OC_NAMESPACE}")
    # shellcheck disable=SC2181
    if [ "$?" -eq 0 ]; then
        # Check pod status. it must be 1/1
        if grep -q "1/1" <<< "$agentStatus"; then
            echo "The agent $BRIDGE_AGENT_NAME is running"
            break
        else
            echo "The agent ${BRIDGE_AGENT_NAME} is not yet running. Will check the status again in $retryDelay seconds."
            sleep "$retryDelay"
        fi
    else
        if [ $i = maxRetries ] ; then
            echo "The agent ${BRIDGE_AGENT_NAME} failed to start within the timeout period. Analyze agent deployment or pod logs and fix the issue. Script will now end. $agentStatus" 
            exit 1
        fi
    fi
done

# Step 15: Deploy destination agent
echo "Deploying agent $DEST_AGENT_NAME"
ocrc=$(oc apply -f $OC_DEST_AGENT_DEPLOYMENT -n "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ "$?" -ne 0 ] ; then
    # shellcheck disable=SC2319
    echo "Failed to deploy bridge agent $OC_DEST_AGENT. Command failed with return code $?. Script will now end. $ocrc"
    exit 1
fi

# Step 16: Verify the destination agent pod is up and running
echo "Verifying status of agent $DEST_AGENT_NAME"
for (( i=1; i <= maxRetries; ++i ))
do
    agentStatus=$(oc get deployment "${OC_DEST_AGENT}" -n "${OC_NAMESPACE}")
    # shellcheck disable=SC2181
    if [ "$?" -eq 0 ]; then
        # Check pod status. it must be 1/1
        if grep -q "1/1" <<< "$agentStatus"; then
            echo "The agent $DEST_AGENT_NAME is running"
            break
        else
            echo "The agent ${DEST_AGENT_NAME} is not yet running. Will check the status again in $retryDelay seconds."
            sleep "$retryDelay"
        fi
    else
        if [ $i = maxRetries ] ; then
            echo "The agent ${DEST_AGENT_NAME} failed to start within the timeout period. Analyze agent deployment or pod logs and fix the issue. Script will now end. $agentStatus" 
            exit 1
        fi
    fi
done

echo "IBM MQ Managed File Transfer agent deployment is now complete."


# Step 17: Run transfers 
echo "Running a file transfer between $SRC_AGENT_NAME and $DEST_AGENT_NAME"
podSrc=$(oc get pods -n "${OC_NAMESPACE}" -l app="${OC_SRC_AGENT}" -o jsonpath='{.items[0].metadata.name}')
 # shellcheck disable=SC2181
 if [ "$?" -eq 0 ]; then
    echo "Creating a file in source agent "
    crtFile=$(oc exec "$podSrc" -it -- cp notices.txt mountpath/hello.txt )
    if [ "$?" -eq 0 ]; then
        echo "Submitting transfer request in source agent pod $podSrc"
        crtTrns=$(oc exec "$podSrc" -it -- fteCreateTransfer -w 30 -de overwrite -da $DEST_AGENT_NAME -dm $QM_NAME -sa $SRC_AGENT_NAME -sm $QM_NAME -df "/mountpath/hello.txt" "/mountpath/hello.txt" )
        if [ "$?" -eq 0 ]; then
            echo "Transfer request completed successfully. $crtTrns"
        else
            echo "Failed to submi transfer request. $crtTrns"
        fi
    else
        echo "Failed to create a file on source agent $SRC_AGENT_NAME due to errir $crtFile" 
    fi
 fi
