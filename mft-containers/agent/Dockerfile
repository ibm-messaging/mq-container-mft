# This docker file builds an IBM MQ Managed File Transfer Agent based on RHEL UBI 
FROM registry.access.redhat.com/ubi8/ubi:8.1

# Add the Agent redistributable package
COPY 9.2.0.0-IBM-MQFA-Redist-LinuxX64.tar.gz /mftdkr/9.2.0.0-IBM-MQFA-Redist-LinuxX64.tar.gz
COPY *.sh /mftdkr/

# Add the required XML files
COPY *.xml /usr/local/bin/
COPY monitor /mftdkr/monitor/

# Update current packages 
RUN yum -y update

# Get EPEL
# RUN rpm -iUvh http://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm

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
# C Runtime
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
#  epel-release \
#  inotify-tools \
  glibc-all-langpacks \
#  s3fs - not available on UBI \
# Download and extract the MQ installation files
  && mkdir -p /var/mqm/mft \
  && useradd samantha \
  && echo samantha:H0neycomb | chpasswd \
  && mv /mftdkr/*.sh /usr/local/bin/  \
  && chmod +x /usr/local/bin/*.sh \
  && mv /mftdkr/*.gz /var/mqm/mft/ \
  && cd /var/mqm/mft \
  && tar -zxvf ./*.tar.gz \
  && yum update \
  # Remove tar.gz files unpacked by RPM postinst scripts
  && find /var/mqm/mft -name '*.tar.gz' -delete \
  # Apply any bug fixes not included in base Ubuntu or MQ image.
  # Don't upgrade everything based on Docker best practices https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/#run
#  && yum -y install krb5-server krb5-libs \
#  && yum update libexpat1 \
  # End of bug fixes
  && rm -rf /var/lib/apt/lists/*
#  && localedef -f UTF-8 -i en_US ./en_US.UTF-8 
#  && localedef 
# Call our entry point
ENTRYPOINT ["mqft.sh"]
