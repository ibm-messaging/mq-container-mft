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

# Queue Manager deployment Name
OC_QM_NAME=mftqm
# Bridge Agent deployment name
OC_BRIDGE_AGENT="bridge-agent"
# Source Agent deployment name
OC_SRC_AGENT="source-agent"
# Destination agent deployment name
OC_DEST_AGENT="dest-agent"
# Source agent PVC
PVC_SOURCE="pvc-source"
# Destination agent PVC
PVC_DEST="pvc-dest"
# Queue manager configMap
QM_CONFIG_MAP=agent-qm-configmap
# Source agent configMap
SOURCE_AGENT_CONFIG_MAP=src-agent-config
# Bridge agent configMap
BRIDGE_AGENT_CONFIG_MAP=pba-config
# Bridge agent credentials configMap
BRIDGE_AGENT_CRED_CONFIG_MAP=pba-custom-cred
# Destination agent configMap
DEST_AGENT_CONFIG_MAP=dest-agent-config

# STEP 1: Switch to ibm-mqmft namespace
ocrc=$(oc project "${OC_NAMESPACE}" > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -gt 0 ] ; then
    echo "ibm-mqmft namespace may not exist. Command failed with return code $ocrc Script will end now.$ocrc"
    exit 1
fi

# STEP2: Delete source agent
echo "Deleting agent $OC_SRC_AGENT"
ocrc=$(oc get deployment $OC_SRC_AGENT > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete deployment $OC_SRC_AGENT > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "Agent $OC_SRC_AGENT deleted"
    else
        echo "An error $ocrc occurred while trying to delete agent $OC_SRC_AGENT"
    fi
else
    echo "Agent deployment $OC_SRC_AGENT not found."
fi

# STEP3: Delete destination agent
echo "Deleting agent $OC_DEST_AGENT"
ocrc=$(oc get deployment $OC_DEST_AGENT > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete deployment $OC_DEST_AGENT > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "Agent $OC_DEST_AGENT deleted"
    else
        echo "An error $ocrc occurred while trying to delete agent $OC_DEST_AGENT"
    fi
else
    echo "Agent deployment $OC_DEST_AGENT not found."
fi

# STEP4: Delete bridge agent
echo "Deleting agent $OC_BRIDGE_AGENT"
ocrc=$(oc get deployment $OC_BRIDGE_AGENT > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete deployment $OC_BRIDGE_AGENT > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "Agent $OC_BRIDGE_AGENT deleted"
    else
        echo "An error $ocrc occurred while trying to delete agent $OC_BRIDGE_AGENT"
    fi
else
    echo "Agent $OC_BRIDGE_AGENT deployment not found."
fi

# STEP5: Delete bridge agent configmap
echo "Deleting ConfigMap $BRIDGE_AGENT_CONFIG_MAP"
ocrc=$(oc get configmap $BRIDGE_AGENT_CONFIG_MAP > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete configmap $BRIDGE_AGENT_CONFIG_MAP > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$BRIDGE_AGENT_CONFIG_MAP deleted"
    else
        echo "An error $ocrc occurred while trying to delete agent configmap $BRIDGE_AGENT_CONFIG_MAP"
    fi
else
    echo "Configmap $BRIDGE_AGENT_CONFIG_MAP not found."
fi

# STEP6: Delete bridge agent credentials configmap
echo "Deleting ConfigMap $BRIDGE_AGENT_CRED_CONFIG_MAP"
ocrc=$(oc get configmap $BRIDGE_AGENT_CRED_CONFIG_MAP > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete configmap $BRIDGE_AGENT_CRED_CONFIG_MAP > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$BRIDGE_AGENT_CRED_CONFIG_MAP deleted"
    else
        echo "An error $ocrc occurred while trying to delete agent configmap $BRIDGE_AGENT_CRED_CONFIG_MAP"
    fi
else
    echo "Configmap $BRIDGE_AGENT_CRED_CONFIG_MAP not found."
fi

# STEP7: Delete source agent configmap
echo "Deleting ConfigMap $SOURCE_AGENT_CONFIG_MAP"
ocrc=$(oc get configmap $SOURCE_AGENT_CONFIG_MAP > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete configmap $SOURCE_AGENT_CONFIG_MAP > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$SOURCE_AGENT_CONFIG_MAP deleted"
    else
        echo "An error $ocrc occurred while trying to delete agent configmap $SOURCE_AGENT_CONFIG_MAP"
    fi
else
    echo "Configmap $SOURCE_AGENT_CONFIG_MAP not found."
fi

# STEP8: Delete destination agent configmap
echo "Deleting ConfigMap $DEST_AGENT_CONFIG_MAP"
ocrc=$(oc get configmap $DEST_AGENT_CONFIG_MAP > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete configmap $DEST_AGENT_CONFIG_MAP > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$DEST_AGENT_CONFIG_MAP deleted"
    else
        echo "An error $ocrc occurred while trying to delete agent configmap $DEST_AGENT_CONFIG_MAP"
    fi
else
    echo "Configmap $DEST_AGENT_CONFIG_MAP not found."
fi

# STEP9: Delete Destination Persistent Volume Claims
echo "Deleting Persistent Volume Claim $PVC_DEST"
ocrc=$(oc get pvc $PVC_DEST > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete pvc $PVC_DEST > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$PVC_DEST deleted"
    else
        echo "An error $ocrc occurred while trying to Persistent Volume Claim $PVC_DEST"
    fi
else
    echo "Persistent Volume Claim $PVC_DEST not found."
fi

# STEP10: Delete Source Persistent Volume Claims
echo "Deleting Persistent Volume Claim $PVC_SOURCE"
ocrc=$(oc get pvc $PVC_SOURCE > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete pvc $PVC_SOURCE > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$PVC_SOURCE deleted"
    else
        echo "An error $ocrc occurred while trying to delete Persistent Volume Claim $PVC_SOURCE"
    fi
else
    echo "Persistent Volume Claim $PVC_SOURCE not found."
fi

# STEP11: Delete queue manager configmap
echo "Deleting ConfigMap $QM_CONFIG_MAP"
ocrc=$(oc get configmap $QM_CONFIG_MAP > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete configmap $QM_CONFIG_MAP > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$QM_CONFIG_MAP deleted"
    else
        echo "An error $ocrc occurred while trying to delete queue manager configmap $QM_CONFIG_MAP"
    fi
else
    echo "Configmap $QM_CONFIG_MAP not found."
fi


# STEP12: Delete queue manager 
echo "Deleting queue manager $OC_QM_NAME"
ocrc=$(oc get queuemanager $OC_QM_NAME > /dev/null 2>&1)
# shellcheck disable=SC2181
if [ $? -eq 0 ] ; then
    ocrc=$(oc delete queuemanager $OC_QM_NAME > /dev/null 2>&1)
    if [ $? -eq 0 ]; then
        echo "$OC_QM_NAME deleted"
    else
        echo "An error $ocrc occurred while trying to delete queue manager $OC_QM_NAME"
    fi
else
    echo "Queue manager $OC_QM_NAME not found."
fi

