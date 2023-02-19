# IBM MQ Managed File Transfer Agent in Container

## Overview
IBM MQ Managed File Transfer transfers files between systems in a managed and auditable way, regardless of file size or the operating systems used. You can use Managed File Transfer to build a customized, scalable, and automated solution that enables you to manage, trust, and secure file transfers. Managed File Transfer eliminates costly redundancies, lowers maintenance costs, and maximizes your existing IT investments.

Run IBM MQ Managed File Transfer agents in a container.

See [here](archive/README.md) for an earlier implementation of MFT on cloud.

## What is new in this version
**9.3.2.0**
- Container image built with 9.3.2.0 CD of IBM MQ Managed File Transfer Redistributable Image.
- Fixes issues found in internal testing and by customers.

**9.3.1.0**
This version of container image supports TLS secure connections to queue managers. You can now specify cipherspec environment variables as described below. The public keys must be mounted into the container at a specific path. See [here](docs/tls.md) for more details.

Starting MQ Version 9.2.5, agents can write progress of transfers to a file in JSON format to agent's log directory. This version of agent container image can publish the contents of transfer logs to a logDNA server. Connection information of logDNA server can be supplied through environment variable **MFT_TLOG_PUBLISH_INFO**. The environment variable must point to JSON formatted file containing logDNA server connection details. The format of the JSON file is described [here](docs/tlogpublsh.md)

## Developer image
Developer version of the MFT Agent container image is available in IBM Container Registry `(icr.io/ibm-messaging/mqmft)`. Use podman/docker command to pull the image.

`podman pull icr.io/ibm-messaging/mqmft`


## Usage

See [here](docs/usage-podman.md) for details on how to run the image container with Podman runtime. 

See [here](docs/usage-ocp.md) for details on how to deploy the image in an OpenShift Container Platform.


Note that in order to use the image, it is necessary to accept the terms of the [IBM MQ license](#license).

### Environment variables supported by this image

- **LICENSE** - Required. Set this to `accept` to agree to the MQ Advanced for Developers license. If you wish to see the license you can set this to `view`.
- **MFT_AGENT_CONFIG_FILE** - Required. Path of the json file containing information required for setting up an agent. The path must be on a mount point. For example a configMap on OpenShift. See the [agent configuration doc](docs/agentconfig.md) for a detailed description of attributes.
- **MFT_AGENT_NAME** - Required. Name of the agent to configure. 
- **BFG_JVM_PROPERTIES** - Optional - Any JVM property that needs to be set when running agent JVM.
- **MFT_LOG_LEVEL** - Optional - Level of information displayed. `info` and `verbose` are the supported values with `info` being default. Contents of agent's output0.log is displayed if MFT_LOG_LEVEL is set to `verbose`.
- **MFT_AGENT_START_WAIT_TIME** - Optionl. An agent might take some time to start after fteStartAgent command is issued. This is the time, in seconds, the containor will wait for an agent to start. If an agent does not within the specified wait time, the container will end.
- **MFT_MOUNT_PATH** - Optional. Environment variable pointing to path from where agent will read files or write to.
- **MFT_COORD_QMGR_CIPHER** - Name of the CipherSpec to be used for securely connecting to coordination queue manager. 
- **MFT_CMD_QMGR_CIPHER** - Name of the CipherSpec to be used for securely connecting to command queue manager. 
- **MFT_AGENT_QMGR_CIPHER** -Name of the CipherSpec to be used for securely connecting to agent queue manager. 

### Location of agent configuration files

Agent in the container will create agent configuration and log files under the fixed directory `/mnt/mftdata`. This folder can be on a persistent volume as well, in which case the volume must be mounted as `/mnt/mftdata` mount point in to the container

### Building your own container image
See the instructions [here](docs/build.md) to build your own agent container image.

### Lab 
Step-by-step [guide](lab/README.md) to using agent container.

### Limitations
1) Private key and trust stores are not yet supported for Protocol Bridge Agents. Hence only a userid and password combination must be used for connecting to FTP/SFTP/FTPS servers.

## Issues and contributions
### Known issues

When using secure connections to queue manager, agent running in a container may log the following warning messages to console or agent's output0.log. Container will continue to run though.
```
[11/07/2022 07:32:16:099 GMT] 00000022 FileSystemPre W   Could not lock User prefs.  Unix error code 2.
[11/07/2022 07:32:16:100 GMT] 00000022 FileSystemPre W   Couldn't flush user prefs: java.util.prefs.BackingStoreException: Couldn't get file lock.

```

Do the following to resolve the warnings:
1) Include the following environment variable in your deployment yaml if you are deploying in OpenShift Container Platform
 ```
 - name: BFG_JVM_PROPERTIES
   value: -Djava.util.prefs.systemRoot=/jprefs/.java/.systemPrefs -Djava.util.prefs.userRoot=/jprefs/.java/.userPrefs

```
2) Include the following environemt variable while running podman/docker runtime:
```   
  --env BFG_JVM_PROPERTIES=-Djava.util.prefs.systemRoot=/jprefs/.java/.systemPrefs -Djava.util.prefs.userRoot=/jprefs/.java/.userPrefs
```
   

For issues relating specifically to the container image, please use the [GitHub issue tracker](https://github.com/ibm-messaging/mft-cloud/issues). If you do submit a Pull Request related to this container image, please indicate in the Pull Request that you accept and agree to be bound by the terms of the [IBM Contributor License Agreement](CLA.md).

## Licenses

The Dockerfiles and associated code and scripts are licensed under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
Licenses for the products installed within the images are as follows:

- [IBM MQ Advanced for Developers](http://www14.software.ibm.com/cgi-bin/weblap/lap.pl?la_formnum=Z125-3301-14&li_formnum=L-APIG-BMKG5H) (International License Agreement for Non-Warranted Programs). This license may be viewed from an image using the `LICENSE=view` environment variable as described above or by following the link above.
- [IBM MQ Advanced](http://www14.software.ibm.com/cgi-bin/weblap/lap.pl?la_formnum=Z125-3301-14&li_formnum=L-APIG-BMJJBM) (International Program License Agreement). This license may be viewed from an image using the `LICENSE=view` environment variable as described above or by following the link above.

Note: The IBM MQ Advanced for Developers license does not permit further distribution and the terms restrict usage to a developer machine.

## Copyright

© Copyright IBM Corporation 2020, 2023
