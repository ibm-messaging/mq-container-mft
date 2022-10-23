# Issues and contributions

Starting with 9.2.4 CD release, IBM MQ Managed File Transfer creates additional logs to track the progress of a transfer more granularly. The logs, in json format, are written to transferlog0.json file under the agent's log directory. The MFT container image provides an option using which transfer logs can be automatically pushed to logDNA server. The details of the logDNA server like the URL, injestion key must be provided as a JSON object through a kubernetes secret. The contents of the secret must be Base64 encoded. Here is the JSON structure:

```
{ 
	"type":"logDNA", 
	"logDNA":{
		"url":"https://<your logdna host name>/logs/ingest",
		"injestionKey":"<your injestion key>"
	}
}
```

Example of sample secret is given below

```
  kind: Secret
  apiVersion: v1
  metadata:
    name: logdna-secret
    namespace: ibmmqft
  data:
    logdna.json: >-
      <base64 encoded data>
```

The secret must be mounted into the container through. A snippet of deployment yaml is here:

```
    spec:
      volumes:
        - name: logdna-url-secret
          secret:
            secretName: logdna-secret
            items:
            - key: logdna.json
              path: logdna.json
      env:
        - name: MFT_AGENT_TRANSFER_LOG_PUBLISH_CONFIG_FILE
          value: /logdna/logdna.json
      volumeMounts:
        - name: logdna-url-secret
          mountPath: /logdna/logdna.json
          subPath: logdna.json
          readOnly: true              
```
