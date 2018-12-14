---
copyright:
  years: 2017, 2018
lastupdated: "2018-12-14"
---

{:new_window: target="_blank"}
{:shortdesc: .shortdesc}
{:screen: .screen}
{:codeblock: .codeblock}
{:pre: .pre}

# Running MQ Managed File Transfer in Kubernetes
{: #mft_intro}

[MFT-Containers]() explains how to run MFT Agents in a container. This document guides you further on how to run the MFT Agent containers in your Kubernetes cluster on IBM Cloud platform.

{:shortdesc}
---

## Prerequisites
{: #mft_containers_prereq}

1. Complete following excercise [MFT-Containers](https://github.com/darbhakiran/mqmft-container)
2. [Create Kubernetes cluster](https://console.bluemix.net/docs/containers/container_index.html#container_index)  
**Note:**
    1. For the pre-requisite, create Kubernetes cluster only. For now, you can ignore **Whats Next** section in the above page.
    2. This turorial creates a free Kubernetes cluster. There paid options that offer more capabilities may also be considered.
3. Install [IBM Cloud CLI & Contianer Registry Plugin](https://console.bluemix.net/docs/cli/index.html#overview)

---

### Push MQ and MFT images to IBM Cloud Docker Registry

Please follow below steps to push customized MQ image and the newly created MFT images into IBM Cloud Docker registry. Inorder to deploy the mq and mft images into a kubernetes cluster, we will first push them into docker registry and create containers of them using deployment/helm scripts.

##### Setup a namespace
Create a namespace to push images

1. Open a command-shell and log in to IBM Cloud.
    ```
    ibmcloud login
    ```
2. Add a namespace to create your own image repository. Replace <my_namespace> with your preferred namespace.
    ```
    ibmcloud cr namespace-add <my_namespace>
    ```
    `ex: ibmcloud cr namespace-add mft-images`

3. To ensure that your namespace is created, run the ibmcloud cr namespace-list command.
    ```
    ibmcloud cr namespace-list
    ```

##### Tag your images
Tagging your images allows you to version and helps in easy maintainance of different versions of it.  

```  
docker tag <source_image>:<tag> registry.<region>.bluemix.net/<my_namespace>/<new_image_repo>:<new_tag>  
```  

**Note:**  
1. <source_image>:<tag> - Is the image you want to tag. Use exact name and tag of the docker image as you see in your local environment. You can find this by running command `docker images` on your local environment.
2. Replace <region> with the name of your region. Replace <my_namespace> with the namespace that you created in **Setup a namespace** section.   
3. Replace <new_image_repo> and <new_tag> with the target image repository name and chose a tag for it.  
4. Run the command for mqadvanced image and also for the agent redistributable image.  
5. For example:  
- `docker tag mqadvmft:latest registry.eu-de.bluemix.net/mft-images/mqadvmft:1.0`  
- `docker tag mftagentredist:latest registry.eu-de.bluemix.net/mft-images/agentredist_mftimagerepo:1.0`  
                  

##### Push the tagged docker images to your cloud regitry
In order to deploy containers of the docker images of the mqadvanced and mft redistributable agent, we need to push the tagged docker images into cloud docker registry. 
```
docker push registry.<region>.bluemix.net/<namespace>/<image>:<tag>
```
For example:
- `docker push registry.eu-de.bluemix.net/mft-images/mqadvmft:1.0`
- `docker push registry.eu-de.bluemix.net/mft-images/agentredist_mftimagerepo:1.0`

**Note:** Please do not close the command-shell as the same would be required in steps below.

### Configure kubectl to access Kubernetes cluster
You will need to configure kubectl to access the kubernetes cluster and to carry out the actions to deploy/install the components on this cluster. 
1. Run following command to list the kubernetes clusters in your account
    ```
    ibmcloud cs clusters
    ```
    Sample output:
    ```
    root@trims1:/home/mft-containers-k8# ibmcloud cs clusters
    OK
    Name              ID                                 State    Created      Workers   Location   Version
    mft-kubeCluster   2b8e314339414bc89b369972d9132332   normal   6 days ago   1         mil01      1.10.7_1520
    ```
2.  Run following command to get KUBECONFIG for your cluster
    ```
    ibmcloud cs cluster-config <cluster-name>
    ```
    Sample output
    ```
    OK
    The configuration for mft-kubeCluster was downloaded successfully. Export environment variables to start using Kubernetes.

    export KUBECONFIG=/root/.bluemix/plugins/container-service/clusters/mft-kubeCluster/kube-config-mil01-mft-kubeCluster.yml
    ```
3. Run the `export KUBECONFIG` command available in command output in above step. For example:
    ```
    export KUBECONFIG=/root/.bluemix/plugins/container-service/clusters/mft-kubeCluster/kube-config-mil01-mft-kubeCluster.yml
    ```
4. Run the command `kubectl get pods` to check if the kubectl is configured to access your cluster. If the output shows `No resources found`, then thats expected as there are no resources defined yet and the kubectl is able to access your cluster.

### Queue manager deployment on to Kubernetes cluster
Create a new deployment file(mqadvdeployment.yaml) for deploying the queue manager as a pod and container within it.
```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: mqadvmft-deployment
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mqadvmft
    spec:
      containers:
      - name: mqadvmft
        image: registry.eu-de.bluemix.net/mft-images/mqadvmft:1.0
        env:
        - name: LICENSE
          value: "accept"
        - name: MQ_QMGR_NAME
          value: "QM1"
```  

**Note:** Eye catchers from the file are
1. `image: registry.eu-de.bluemix.net/mft-images/mqadvmft:1.0` - This is the image we tagged and pushed into cloud registry as part of section **Push MQ-MFT images to IBM Cloud Docker Registry**
2. We set number of replicas to 1, this can be increased, for ex:3, by that Kubernetes run three replicas of queue manager pod. If the High-Availability is a required, **replicas** must be atleast 2.
3. A copy of this file is provided as part of the repository files which you can download and use. This is a sample yaml file and can be modified as per your environment.
4. More about this can be found at [Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#creating-a-deployment)

###### Apply the deployment on your Kubernetes cluster
1. Run following command to create the queue manager deployment
    ```
    kubectl create -f filepath/mqadvdeployment.yaml
    ```
2. Check if the deployment is successful and pod is in running state. Run the below command with few seconds gap until you find the pod status to be running
    ```
    kubectl get pods
    ```
3. If incase pods goes into error state. Review the logs to investigate and debug on error cause.
    ```
    kubectl describe pod <pod-name>
    kubectl logs <pod-name> -c <container-name>
    ```
    **Note:** 
    1. `<pod-name>` can be found in the output of command `kubectl get pods`
    2. `<container-name>` can be found in the mqadvdeployment.yaml file.

### Expose the Queue manager ports
Use Kubernetes configuration to make your container accessible over the public internet. For development and test purposes the simplest approach is to create a Kubernetes service that exposes the necessary endpoints using a NodePort, and since our free cluster only has one worker we don’t have to worry about the worker IP address changing.

The following commands create a service for each of the two ports that we want to access in our container;

```
kubectl expose pod <podname> --port 1414 --name mqchannel --type NodePort
kubectl expose pod <podname> --port 9443 --name mqwebconsole --type NodePort
```

Having created the service you now need to look up the port numbers that have been allocated to the NodePort using the **"kubectl get services"** command. In the example below the MQ Channel is exposed publicly on port 30063 and the MQ Web Console on port 32075.

```
kubectl get services

NAME           CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
kubernetes     10.10.10.1     none          443/TCP          22h
mqwebconsole   10.10.10.128   nodes         9443:32075/TCP   2m
mqchannel      10.10.10.44    nodes         1414:30063/TCP   2m
````

Lastly you need to obtain the public IP address of the worker node, for example in the example shown below the public IP address of the worker is 169.51.10.240;

```
bx cs workers <cluster-name>

ID                                                 Public IP       Private IP       Machine Type   State    Status   
kube-par01-pa7f800000007845aaaaf806224d5a53dc-w1   169.51.10.240   10.126.110.230   free           normal   Ready
```

Combine the IP address and the port number together to access the relevant endpoint over the internet, for example;
```
MQ Web Console: https://169.51.10.240:32075/ibmmq/console/ (admin / passw0rd)
MQ Explorer: 169.51.10.240, port 30063 (admin / passw0rd, channel=DEV.ADMIN.SVRCONN)
```
Note: Do not close the command shell as we use the same command shell for next section.

### Configure the Queue manager for MFT
This tutorial uses single queue manager topology, where a single queue manager is configured for MFT Coordination,Command and Agent Queue manager setup.

1. Setup the Queue manager (ex:**QM1**) as coordination queue manager. Running the **mft_setupCoordination.sh** script will create required configuration.

    ```
    kubectl exec -ti <podname> /etc/mqm/mft/mqft_setupCoordination.sh
    ```
2. We will create a source agent on kubernetes cluster as part of this tutorial. These agents are **AGENTSRC**. As a first step of agent configuration, we have to create their congfiguration on the coordination queue manager.

    ```
    kubectl exec -ti <podname> /etc/mqm/mft/mqft_setupAgent.sh AGENTSRC
    ```
    **Note:** 
    1. `<PODNAME>` is the name of the queue manager pod, created in above section.
    2. **mqft_setupAgent.sh** script requires MFT agent name as input parameter

    
### MFT Agents deployment on to Kubernetes cluster

Create a new deployment file(mft_agentredit_Deployment-src.yaml) for deploying the MFT Agent as a pod and container within it.

```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: mft-kube-deployment-src
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mqft-kube-src
    spec:
      containers:
      - name: agentsrc
        image: registry.eu-de.bluemix.net/mft-images/agentredist_mftimagerepo:1.0
        env:
        - name: MQ_QMGR_NAME
          value: "QM1"
        - name: MQ_QMGR_HOST
          value: "169.51.10.240"
        - name: MQ_QMGR_PORT
          value: "32075"
        - name: MQ_QMGR_CHL
          value: "MFT.SVRCONN"
        - name: MFT_AGENT_NAME
          value: "AGENTSRC"
```

**Note:** Eye catchers from the file are  
1. `image:  registry.eu-de.bluemix.net/mft-images/agentredist_mftimagerepo:1.0` - This is the image we tagged and pushed into cloud registry as part of section **Push MQ-MFT images to IBM Cloud Docker Registry**
2. We set number of replicas to 1, this can be increased, for ex:3, by that Kubernetes run three replicas of queue manager pod. If the High-Availability is a required, **replicas** must be atleast 2.
3. `-env` parameters **MQ_QMGR_HOST** AND **MQ_QMGR_PORT** refer to public-ip and exposed port of the qmgr pod in your kubernetes cluster. Refer to **Expose the Queue manager ports** above.
4. More about this can be found at [Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#creating-a-deployment)

Ideal approach is to create each agent as deployment. Above deployment file(mft_agentredit_Deployment-src.yaml) is for one agent, this turtorial used it as Source Agent.
In order to run it for destination agent, 
1. Create a new file by name "mft_agentredit_Deployment-dest.yaml", 
2. Copy the contents of file "mft_agentredit_Deployment-src.yaml" into it 
3. Find and replace **src** with **dest**.
4. All the other settings would remain same for destination agent.

A copy of these deployment files is provided as part of the repository files, which you can download and use. These files are sample files and can be updated based on your environment.

###### Apply the deployment on your Kubernetes cluster
1. Run following command to create the MFT Agent deployment
    ```
    kubectl create -f filepath/mft_agentredit_Deployment-src.yaml
    ```
2. Check if the deployment is successful and pod is in running state. Run the below command with few seconds gap until you find the pod status to be running
    ```
    kubectl get pods
    ```
3. If incase pods goes into error state. Review the logs to investigate and debug on error cause.
    ```
    kubectl describe pod <pod-name>
    kubectl logs <pod-name> -c <container-name>
    ```
    **Note:** 
    1. `<pod-name>` can be found in the output of command `kubectl get pods`
    2. `<container-name>` can be found in the mqadvdeployment.yaml file(at spec-->containers-->name)
4. Repeat steps 1-3 for destination agent deployment using the **mft_agentredit_Deployment-dest.yaml** file.

### Verify if the MFT Agents are configured correctly.
Inorder to verify that MFT environment and its Agents are configured correctly,we will run the **fteListAgents** command that will list currently available agents.
1. On the command-shell, run following command to list all the pods
    ```
    kubectl get pods
    ```
2. Using the any one of the MFT Agent's pod-name available from above output, run below command
    ```
    kubectl exec -ti <podname> fteListAgents
    ```
    Sample output
    ```
    root@trims1:/home/mft-containers-k8# kubectl exec -ti mft-kube-deployment-dest-c66c55567-pdvtn fteListAgents
    5724-H72 Copyright IBM Corp.  2008, 2018.  ALL RIGHTS RESERVED
    Command executed at 2018-09-18 09:53:15 UTC
    Coordination queue manager time 2018-09-18 09:53:15 UTC
    Agent Name:     Queue Manager Name:     Status:     Status Age:
    AGENTDEST       QM1                     READY           0:03:25
    AGENTSRC        QM1                     READY           0:04:42
    ```
If the command output in your environment is similar to the above Sample output, that confirms that MFT Agents are configured and running in your kubernetes cluster.

### Conclusion 
As part of this document we have pushed docker images into cloud registry, created a kubernetes cluster, deployed mq and mft-agents containers onto kubernetes cluster and verified agents are configured correctly and responding to mft commands.

#### Related information
This article explains on running IBM MQ DockerHub image on your kubernetes  
[Running the MQ docker image on the Kubernetes service in Bluemix](https://developer.ibm.com/messaging/2017/09/04/kubernetes-service-mq-docker-bluemix/)
