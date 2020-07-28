# This docker file builds an IBM MQ Managed File Transfer Agent based on RHEL UBI 
FROM registry.access.redhat.com/ubi8/ubi:8.1

# Add the Agent redistributable package
COPY 9.2.0.0-IBM-MQFA-Redist-LinuxX64.tar.gz /mftdkr/9.2.0.0-IBM-MQFA-Redist-LinuxX64.tar.gz
#COPY *.sh /mftdkr/

# Copy golang package
COPY go1.14.6.linux-amd64.tar.gz /golang/go1.14.6.linux-amd64.tar.gz

# copy Golang app to monitor agent
COPY /golang/mftcfg /mftdkr

# Add the required XML files
COPY *.xml /usr/local/bin/

# Update current packages 
RUN yum -y update

# Install additional packages
RUN yum --assumeyes --allowerasing install bash \
  bc \
  ca-certificates \
  coreutils \
  curl \
  file \
  findutils \
  gawk \
  grep \
# Get C Runtime
  glibc \
# check Redhat's system version
  redhat-lsb-core \
# mount drive
  util-linux-ng \  
  passwd \
  procps \
  sed \
  tar \
  util-linux \
  glibc-all-langpacks \
# Download and extract the MQ installation files
  && mkdir -p /var/mqm/mft \
  && useradd samantha \
  && echo samantha:H0neycomb | chpasswd \
#  && mv /mftdkr/*.sh /usr/local/bin/  \
#  && chmod +x /usr/local/bin/*.sh \
  && mv /mftdkr/*.gz /var/mqm/mft/ \
  && cd /var/mqm/mft \
  && tar -zxvf ./9.*.tar.gz \
  && cd /golang \
  && tar -zxvf ./go*.tar.gz \
  && yum update \
  # Remove tar.gz files unpacked by RPM postinst scripts
  && find /var/mqm/mft -name '*.tar.gz' -delete \
  && find /golang -name '*.tar.gz' -delete \
  # Apply any bug fixes not included in base Ubuntu or MQ image.
  # Don't upgrade everything based on Docker best practices https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/#run
  # End of bug fixes
  && rm -rf /var/lib/apt/lists/* 
# Call our entry point
ENV GOPATH=/usr/local/go
ENV PATH=$PATH:/var/mqm/mft/bin:/var/mqm/mft/java/jre64/jre/bin:$GOPATH/bin
ENV BFG_DATA=${ENV_BFGDATA}
#RUN /usr/local/bin/setpath.sh
RUN cp -f /usr/local/bin/MQMFTCredentials.xml  $HOME
RUN cp -f /usr/local/bin/ProtocolBridgeCredentials.xml $HOME
RUN chmod go-rw $HOME/MQMFTCredentials.xml
RUN chmod go-rw $HOME/ProtocolBridgeCredentials.xml
RUN chmod go-rwx /usr/local/bin/MQMFTCredentials.xml

#ENTRYPOINT ["mqft.sh"]
ENTRYPOINT exec "/mftdkr/mftcfg" ${ENV_AGQMGR} ${ENV_AGQMGRHOST} ${ENV_AGQMGRPORT} ${ENV_AGQMGRCHN} ${ENV_AGNAME} ${ENV_BFGDATA} ${ENV_MON_TIME}
