# Building a container image

## Prerequisites

You need to have the following tools installed:

* [Docker](https://www.docker.com/) V17.06.1 or later, or [Podman](https://podman.io) V1.0 or later

If you are working in the Windows Subsystem for Linux, follow [this guide by Microsoft to set up Docker](https://blogs.msdn.microsoft.com/commandline/2017/12/08/cross-post-wsl-interoperability-with-docker/) first.

You will also need a [Red Hat Account](https://access.redhat.com) to be able to access the Red Hat Registry. 

## Building your MFT container image

This procedure works for building the MQ Managed File Transfer Redistributable package on `amd64` architectures.

1. Clone the GitHub repository to local directory.
2. Login to the Red Hat Registry: `docker login registry.redhat.io` using your Customer Portal credentials.
3. Navigate to directory where `Dockerfile-agent` is located.
4. Download **9.2.2.0-IBM-MQFA-Redist-LinuxX64.tar.gz** or higher from [IBM Fixcentral](https://www.ibm.com/support/fixcentral/) into the current directory.

   **Note:** The redistributable MFT package has to be present in same path as **Dockerfile**.
4. Run the following command to build image

   `podman build -f Dockerfile-agent -t mqmft:9.2.2 --build-arg ARG_MQMFT_REDIST_FILE=9.2.2.0-IBM-MQFA-Redist-LinuxX64.tar.gz`
   
   You can replace the `9.2.2.0-IBM-MQFA-Redist-LinuxX64.tar.gz` with the version of the redistributable package of your choice.

### Building the Bridge Credential Exit - bridgecredexit
The container image also ships a MFT Bridge Agent Credential Exit. This exit is used by a Bridge agent to determine credentials required to connect to a SFTP/FTP server.

To build the exit, you require `com.ibm.wmqfte.com.ibm.wmqfte.exitroutines.api.jar` library. The library is shipped as part of IBM MQ Standard Installation image or MFT Redistributable package. 
You can build the exit using Eclipse IDE with at least Java JDK or with `javac` using
   `javac -cp .:/<path>/com.ibm.wmqfte.com.ibm.wmqfte.exitroutines.api.jar:/<path>/json-20210307 ProtocolBridgeCustomCredentialExit.java`

 and jar the class file
   `jar cmf MANIFEST.MF bridgecredexit ProtocolBridgeCustomCredentialExit.class`

## Installed components

This image includes the core MQ Managed File Transfer Agent, IBM Java Runtime.
