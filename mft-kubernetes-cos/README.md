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

# Running MQ Managed File Transfer using Cloud Object Storage
{: #mft_intro}  

Moving files from on-premise to cloud is one of the fastest spreading usecase for hybrid integration scenarios. MQ MFT with IBM Cloud provides solutions that addresses the use case with reliable and secured file transfer capabilities. MQ MFT provides capabilities to schedule, monitor, automate file transfers with complete audit logging capabilities.

![Image showing file transfer scenario](./scenario.JPG)  

In this tutorial we demonstrate a hybrid integration scenario, where a file is transferred from on-premise to IBM Cloud Object Storage(COS).



{:shortdesc}
---

## Prerequisites
{: #mft_containers_prereq}

1. Complete following excercise 
   - [MFT-Containers](https://github.com/darbhakiran/mqmft-container)
   - [Create Kubernetes cluster](https://console.bluemix.net/docs/containers/container_index.html#container_index)  

2. Configure your Cloud Object Storage service instance as persistent volume for your Kubernetes cluster.  
    **Note:** 
         1. Sample configuration file(pvc.yaml) to use cloud object storage as Peristent Volume Claim is provided as part of this tutorial, you can use this and modify it based on your environment if required.  
         2. Sample configuration file for MFT Agent deployment as a pod with its persitent volume claim binded to persistent volume that referencing your service instance of cloud object storage.  
         3. This tutorial uses static provisioning with regional setup of the cloud object storage.  

    Please follow the topic [Storing data on IBM Cloud Object Storage](https://console.bluemix.net/docs/containers/cs_storage_cos.html)  
    Execute the steps listed under following sections only
    - [Creating your object storage service instance](https://console.bluemix.net/docs/containers/cs_storage_cos.html#create_cos_service)  
    - [Creating a secret for the object storage service credentials](https://console.bluemix.net/docs/containers/cs_storage_cos.html#create_cos_secret)
    - [Installing the IBM Cloud Object Storage plug-in](https://console.bluemix.net/docs/containers/cs_storage_cos.html#install_cos)
    - [Deciding on the object storage configuration](https://console.bluemix.net/docs/containers/cs_storage_cos.html#configure_cos)  
    - [Adding object storage to apps](https://console.bluemix.net/docs/containers/cs_storage_cos.html#add_cos)  
    Note: You can use the **pvc.yaml** and **mftdeployment.yaml** file provided as part of this tutorial. Modify the files based on your environment.

### Create an MFT Agent in your on premise environment
This tutorial demonstrates a file transfer from on-premise to cloud object storage. In order to demonstrate this scenario, we will need following 
1. MQ on Cloud Service
2. An MFT Source agent running on premise
3. An MFT Destination agent running in a kunbernetes cluster on IBM Cloud with its persistent valume binded to a bucket on Cloud Object Storage.

### Creating an MQ on Cloud Queue manager with MFT Source Agent on premise
This can be achieved using [MQ on Cloud tutorial](https://console.bluemix.net/docs/services/mqcloud/mqoc_mft_single_qmgr_topology.html#mqoc_mft_single_qmgr_topology)

### Create IBM Cloud Kubernetes cluster
1. [Create Kubernetes cluster](https://console.bluemix.net/docs/containers/container_index.html#container_index)  
**Note:**
    1. For the pre-requisite, create Kubernetes cluster only. For now, you can ignore **Whats Next** section in the above page.
    2. This turorial creates a free Kubernetes cluster. There paid options that offer more capabilities may also be considered.
2. Install [IBM Cloud CLI & Contianer Registry Plugin](https://console.bluemix.net/docs/cli/index.html#overview)

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

### MFT Agents deployment on to Kubernetes cluster

Create a new deployment file(mftdeployment.yaml) for deploying the MFT Agent as a pod and container within it.

```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: mft-kube-deployment-dest
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mqft-kube-dest
    spec:
      containers:
      - name: agentdest
        image: registry.eu-de.bluemix.net/mft-images/mftagents3fs:1.0
        volumeMounts:
        - name: mftstore
          mountPath: /home/MFTLocalStore
        env:
        - name: MQ_QMGR_NAME
          value: "MFT_QMA"
        - name: MQ_QMGR_HOST
          value: "QMgr Host name"
        - name: MQ_QMGR_PORT
          value: "QMgr Port"
        - name: MQ_QMGR_CHL
          value: "QMgr SVRCONN Channel"
        - name: MFT_AGENT_NAME
          value: "IBMCLOUD_AGENT"
      volumes:
      - name: mftstore
        persistentVolumeClaim:
          claimName: mymftpvc
```

**Note:** Eye catchers from the file are  
1. `image:  registry.eu-de.bluemix.net/mft-images/agentredist_mftimagerepo:1.0` - This is the image we tagged and pushed into cloud registry as part of section **Push MQ-MFT images to IBM Cloud Docker Registry**
2. We set number of replicas to 1, this can be increased, for ex:3, by that Kubernetes run three replicas of queue manager pod. If the High-Availability is a required, **replicas** must be atleast 2.
3. `-env` parameters **MQ_QMGR_HOST** AND **MQ_QMGR_PORT** refer to public-ip and exposed port of the qmgr pod in your kubernetes cluster. Refer to **Expose the Queue manager ports** above.
4. `Volumes` refers to persistent volume claim that we provisioned in **Prerequisites** section.
5. More about this can be found at [Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#creating-a-deployment)

A copy of this deployment file is provided as part of the repository files, it can be updated based on your environment and used to deploy the MFT agents in kubernetes environment.  

### Apply the deployment on your Kubernetes cluster
1. Run following command to create the MFT Agent deployment
    ```
    kubectl create -f filepath/mftdeployment.yaml
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
    Command executed at 2018-09-18 16:21:09 UTC
    Coordination queue manager time 2018-09-18 16:21:09 UTC
    Agent Name:     Queue Manager Name:     Status:     Status Age:
    IBMCLOUD_AGENT  QM1                     READY           0:09:16
    AGENTSRC        QM1                     READY           0:14:42
    ```
If the command output in your environment is similar to the above Sample output, that confirms that MFT Agents are configured and running in your kubernetes cluster.

### Create a File Transfer with MFT Agents to write file to Cloud Object Storage(COS)
{: #mft_containers_filetransfer_demo}

1. Open a new command shell
2. Create a text file on **AGENTSRC** for file transfer.
    ```
    echo 'Test file transfer to Cloud Object Storage' > /tmp/cos.txt
    ```
3. Check that file doesn't exist on IBM Cloud Object Storage.  
    -  Log onto your IBM Cloud Object Storage service instance and open the bucket provisioned for file transfer.
    -  Check there's no *cos.txt* file in the bucket.
4. Run the file transfer command to transfer newly created file (**cos.txt**) to **COS**.
    ```
    fteCreateTransfer -p QM1 -sa AGENTSRC -sm QM1 -da IBMCLOUD_AGENT -dm QM1 -df /tmp/cos.txt /home/MFTLocalStore/cos.txt
    ```
5. Wait a seconds and check if file is transferred to configured bucket on **IBM Cloud Object Storage**.  
    5.1 If the file is not available yet, wait a couple of seconds, refresh the bucket and check.   
    5.2 If `5.1` doesn't help, then check into Agent's log (available in *agents* directory in the BFG_DATA path.)

### Conclusion 
As part of this tutorial, we have provisioned cloud object storage service instance as persistent volume in your kubernetes cluster, used it as PVC for the mft agent deployment and performed a file transfer, where the file is transferred to a bucket on cloud object storage.
