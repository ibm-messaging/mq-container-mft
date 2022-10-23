# Building a container image

## Prerequisites

You need to have the following tools installed:

* [Docker](https://www.docker.com/) V17.06.1 or later, or [Podman](https://podman.io) V1.0 or later

If you are working in the Windows Subsystem for Linux, follow [this guide by Microsoft to set up Docker](https://blogs.msdn.microsoft.com/commandline/2017/12/08/cross-post-wsl-interoperability-with-docker/) first.

You will also need a [Red Hat Account](https://access.redhat.com) to be able to access the Red Hat Registry. 

## Building your MFT container image

This procedure works for building the MQ Managed File Transfer Redistributable package on `amd64` architectures.

1. Clone the GitHub repository to local directory.
2. Login to the Red Hat Registry: `podman login registry.redhat.io` using your Customer Portal credentials.
3. Navigate to directory where `Dockerfile-agent` is located.
4. Download **9.3.1.0-IBM-MQFA-Redist-LinuxX64.tar.gz** or higher from [IBM Fixcentral](https://www.ibm.com/support/fixcentral/) into the current directory.
5. Unpack the **9.3.1.0-IBM-MQFA-Redist-LinuxX64.tar.gz** to a temporay directory. Copy com.ibm.wmqfte.com.ibm.wmqfte.exitroutines.api.jar to credentialsexit/BridgeCredentialExit/thirdparty directory.
6. Download json-20210307.jar file from [Maven Repository](https://mvnrepository.com/artifact/org.json/json/20210307) and copy to credentialsexit/BridgeCredentialExit/thirdparty directory.
   **Note:** The redistributable MFT package must be present in same path as the **Dockerfile-agent** file.
7. Run the following command to build container image

   `podman build -f Dockerfile-agent -t mqmft:9.3.1 --build-arg ARG_MQMFT_REDIST_FILE=9.3.1.0-IBM-MQFA-Redist-LinuxX64.tar.gz`
   
   You can replace the `9.2.4.0-IBM-MQFA-Redist-LinuxX64.tar.gz` with the version of the redistributable package of your choice.

## Installed components

This image includes the following components
1. MQ Managed File Transfer Agent - Core MFT product.
2. IBM Java Runtime Environment - IBM JRE.
3. json-20210307.jar - Third party JSON parse.
4. Custom Protocol Bridge Credential Exit bridgecredentialexit.jar
5. mqfts - A command line utility to parse and display contents of capture0.log file.