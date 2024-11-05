# Simple scripts to deploy MFT Agent container and a queue manager on OpenShift Cluster.
- This directory in source. This directory contains two main shell scripts and supporting yaml files.
   - The [ocp-deploy](./ocp-deploy.sh) is the main script that deploys agents and queue manager in a OpenShift cluster. 
   - Run `./ocp-deploy.sh 10 30` - This command will attempt to deploy a queue manager instance, 2 instances of standard agent and one bridge agent instance.
   - Will run a transfer between two standard agents.
   - Run the [ocp-clean](../test/ocp-deploy/ocp-clean.sh) script to clean the deployment.
