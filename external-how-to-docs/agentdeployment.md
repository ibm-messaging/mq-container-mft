# Agent deployment file

A deployment YAML is required to deploy an agent in OpenShift Container Platfrom. The following section describes the required deployment attributes and sample values.

```
kind: Deployment
apiVersion: apps/v1
metadata:
  name: ibm-mq-managed-file-transfer-bridgeagent <- Name of the container
  namespace: ibmmqmft <- Project namespace
spec:
  replicas: 1 <- Number of replicas. The supported maximum value is 1
  selector:
    matchLabels:
      app: ibm-mq-managed-file-transfer-bridgeagent
  template:
    metadata:
      labels:
        app: ibm-mq-managed-file-transfer-bridgeagent
        deploymentconfig: ibm-mq-managed-file-transfer-bridgeagent
    
	spec:
      volumes:
        - name: mqmft-agent-config-map <- Name of volume where agent definition attributes in JSON format are defined
          configMap:
            name: mqmft-agent-config <- Name of the configMap containing the definitions of an agent
            
        - name: mqmft-protocol-bridge-cred-map 
          configMap:
            name:  pba-custom-cred-map <- Name of the ConfigMap containing the credential information required for connecting a SFTP/FTP/FTPS server
            
        - name: mqmft-nfs-config 
          persistentVolumeClaim:
            claimName: nfs-mft-pvc - <- Persistent volume that will contain agent configuration and logs
      
	  containers:
        - resources: {}
          readinessProbe:
            exec:
              command:
                - agentready <- Readiness probe that monitors agent readiness
            initialDelaySeconds: 15
            timeoutSeconds: 3
            periodSeconds: 30
            successThreshold: 1
            failureThreshold: 3
    
   	      terminationMessagePath: /dev/termination-log
          
		  name: ibm-mq-managed-file-transfer-bridgeagent <- Name of the container
          livenessProbe:
            exec:
              command:
                - agentalive <- Liveness probe that monitors agent health
            initialDelaySeconds: 90
            timeoutSeconds: 5
            periodSeconds: 90
            successThreshold: 1
            failureThreshold: 3
			
          env:
            - name: MFT_AGENT_NAME <- Required. Environment variable containing the name of the agent to deploy
              value: BRIDGE
			- name: LICENSE <- Required. Environment variable for accepting license to use an agent
              value: accept
            - name: BFG_DATA <- Optional. Path where configuration and log files are created. 
              value=/mnt/mftdata
			- name: MFT_AGENT_CONFIG_FILE <- Name of environment variable containing the name of the JSON file containing information required to cofingure an agent. The JSON file must reside in a ConfigMap 
              value: /mqmftcfg/agentconfig/mqmftcfg.json <- Path of the JSON file containing agent definitions 
			- name: MFT_BRIDGE_CREDENTIAL_FILE <- Required for BRIDGE agent only. Name of the environment variable that points to path of a file containing credential information for connecting to SFTP/FTP/FTPS file server. The file can reside either in a configMap or secret.
              value: /mqmftbridgecred/agentcreds/ProtocolBridgeCredentials.prop <- Path of the file containing bridge credential information
          - name: MFT_LOG_LEVEL Optional. Controls the amount of debug information displayed while deploying the container. Default is "info"
            value="verbose"
          - name: MFT_TRACE_COMMAND Optional. Enable tracing of MFT commands. Default is "no".
             value="yes"
          - name: MFT_TRACE_COMMAND_PATH Required if MFT_TRACE_COMMAND is set to yes. The path where trace files will be created.
            value: /mnt/mftdata 
          - name: MFT_AGENT_START_WAIT_TIME Optional. Time in seconds to for container to wait for agent to start. Default is 10 seconds. If the agent does not start in the specified time, the container ends.
            value=15
		  imagePullPolicy: Always
          
		  volumeMounts:
            - name: mqmft-agent-config-map <- Mount path where JSON file containing information required to confiure an agent resides
              mountPath: /mqmftcfg/agentconfig
            
			- name: mqmft-protocol-bridge-cred-map <- Mount path where custom credential file for a BRIDGE agent resides.
              mountPath: /mqmftbridgecred/agentcreds
            
			- name: mqmft-nfs-config <-Optional: Mount path where configuration and log files of an agent would be created
              mountPath: /mnt/mftdata
              subPath: mftdata
          
		  terminationMessagePolicy: File
          
		  image: >-
             icr.io/ibm-messaging/mqmft:latest <- URI from where an agent container image will be pulled for deployment
```
