# Protocol Bridge Credential File. 
A bridge agent requires additional information, like the user id and password to connect to external FTP/FTPS/SFTP servers. This information must be provided as key value pairs in a file as described below. The file must be available on a mount point and name of the file is provided as value of `protocolBridgeCredentialConfiguration` attribute in agent's JSON configuration file. For example `"protocolBridgeCredentialConfiguration" : "/mqmftbridgecred/agentcreds/ProtocolBridgeCredentials.prop"`.

A configMap or a secret can also be used when running on OpenShift Container Platform. The credential information must be in the following format.


`<hostName>=<User ID>!<type>!<password>`

Where 
- **hostName** - Host name or IP address of the SFTP/FTP/FTPS server.
- **User ID** - User id for connecting to the SFTP/FTP/FTPS server.
- **Type** - Type of password - plain text or Base64 encoded

             0 - Plain text password
			 
			 1 - Base64 encoded password.

- **Password** - Password of the user for connecting to the SFTP/FTP server.

Example:

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: pba-custom-cred-map
  namespace: ibmmqmft
data:
  ProtocolBridgeCredentials.prop: sftp.server.com=sftpuid!0!SftpPassw0rd
```
### Limitations
1) Private key and trust stores are not yet supported for Protocol Bridge Agents. Hence only a userid and password combination must be used for connecting to FTP/SFTP/FTPS servers.
