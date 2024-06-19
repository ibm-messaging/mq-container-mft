# Protocol Bridge Credential File. 
A bridge agent requires additional information, like the user id and password to connect to external FTP/FTPS/SFTP servers. This information must be provided as key value pairs in a file as described below. The file must be available on a mount point and name of the file is provided as value of `protocolBridgeCredentialConfiguration` attribute in agent's JSON configuration file. For example `"protocolBridgeCredentialConfiguration" : "/mqmftbridgecred/agentcreds/ProtocolBridgeCredentials.prop"`.

A configMap or a secret can also be used when running on OpenShift Container Platform. 

The credential information can be specified in one of the following two formats.

### Key value pair of Hostname-credentials

`<hostName>=<User ID>!<type>!<password>`

Where 
- **hostName** - Host name or IP address of the SFTP/FTP/FTPS server.
- **User ID** - User id for connecting to the SFTP/FTP/FTPS server.
- **Type** - Type of password - plain text or Base64 encoded

   0 - Plain text password

   1 - Base64 encoded password.

- **Password** - Password of the user for connecting to the SFTP/FTP server.

Example:

Specify credentials in a file:

sftp.server.com=sftpuid!0!SftpPassw0rd

Specify confidentials as a OpenShift ConfigMap
```
kind: ConfigMap
apiVersion: v1
metadata:
  name: pba-custom-cred-map
  namespace: ibmmqmft
data:
  ProtocolBridgeCredentials.prop: sftp.server.com=sftpuid!0!SftpPassw0rd
```

### Supplying attributes as a JSON object.
Parameters required by ProtocolBridgeAgent for connecting to SFTP server must be supplied through a JSON object. The following are the supported attributes. 

- **serverHostName** - Host name or the IP address of the file server.
- **transferRequesterId** - User Id to match with incoming transfer requests source agent. Transfer requests that don't match the user Id specified will be rejected by the destinatio agent. Please note that in an OpenShift Cluster, agent container may running under a dynamically created user. Hence you may specify '*' to match all user Ids.
- **serverType** - Type of the file server. SFTP and FTP are the supported values. Default is FTP.
- **serverAssocName** - Name to associate.
- **serverUserId** - User Id for connecting to SFTP file server.
- **serverHostKey** - Host key required for connecting to SFTP file server. Must be in Base64 encoded format.
- **serverPrivateKey** - Private key required for connecting to SFTP file server. Must be in Base64 encoded format.
- **serverPassword** - Password for the private key.

An example configmap:

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: pba-custom-cred-map
data:
  ProtocolBridgeCredentials.prop: {
 "servers": [
	{
	"serverHostName":"sftp.host.com",
	"transferRequesterId":"mquserid",
	"serverType": "SFTP",
	"serverUserId": "sftpuserid",
	"serverAssocName": "<sftpassconame>",
	"serverHostKey": "<base64 encoded host key>",
	"serverPassword": "<password of private key>",
	"serverPrivateKey": "<base64 encoded private key>"
	}
	]
}
```
