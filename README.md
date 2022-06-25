# IBM MQ Managed File Transfer Container

## Overview
IBM MQ Managed File Transfer transfers files between systems in a managed and auditable way, regardless of file size or the operating systems used. You can use Managed File Transfer to build a customized, scalable, and automated solution that enables you to manage, trust, and secure file transfers. Managed File Transfer eliminates costly redundancies, lowers maintenance costs, and maximizes your existing IT investments.

Run IBM MQ Managed File Transfer agents in a container .

See [here](archive/README.md) for an earlier implementation of MFT on cloud.


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
- **MFT_BRIDGE_CREDENTIAL_FILE** - Required for bridge type agents. Path of the user supplied protocol bridge credentials file used for connecting to protocol servers. See [here](docs/custompbacred.md) for more details on the format of the credential file. 
- **MFT_BRIDGE_PROPERTIES_FILE** - Optional. Valid only for bridge type agents. Path of the user supplied bridge properties file in XML format.

### Location of agent configuration files

Agent in the container will create agent configuration and log files under the fixed directory `/mnt/mftdata`. This folder can be on a persistent volume as well, in which case the volume must be mounted as `/mnt/mftdata` mount point in to the container

### Building your own container image
See the instructions [here](docs/build.md) to build your own agent container image.

### Lab 
Step-by-step [guide](lab/README.md) to using agent container.

### Limitations
1) Private key and trust stores are not yet supported for Protocol Bridge Agents. Hence only a userid and password combination must be used for connecting to FTP/SFTP/FTPS servers.

## Issues and contributions

For issues relating specifically to the container image, please use the [GitHub issue tracker](https://github.com/ibm-messaging/mft-cloud/issues). If you do submit a Pull Request related to this container image, please indicate in the Pull Request that you accept and agree to be bound by the terms of the [IBM Contributor License Agreement](CLA.md).

## Licenses

The Dockerfiles and associated code and scripts are licensed under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
Licenses for the products installed within the images are as follows:

- [IBM MQ Advanced for Developers](http://www14.software.ibm.com/cgi-bin/weblap/lap.pl?la_formnum=Z125-3301-14&li_formnum=L-APIG-BMKG5H) (International License Agreement for Non-Warranted Programs). This license may be viewed from an image using the `LICENSE=view` environment variable as described above or by following the link above.
- [IBM MQ Advanced](http://www14.software.ibm.com/cgi-bin/weblap/lap.pl?la_formnum=Z125-3301-14&li_formnum=L-APIG-BMJJBM) (International Program License Agreement). This license may be viewed from an image using the `LICENSE=view` environment variable as described above or by following the link above.

Note: The IBM MQ Advanced for Developers license does not permit further distribution and the terms restrict usage to a developer machine.

## Copyright

© Copyright IBM Corporation 2020, 2022