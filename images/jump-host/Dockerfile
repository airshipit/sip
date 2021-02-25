ARG BASE_IMAGE=gcr.io/gcp-runtimes/ubuntu_18_0_4
FROM ${BASE_IMAGE}

# Inject custom root certificate authorities if needed.
# Docker does not have a good conditional copy statement and requires that a
# source file exists to complete the copy function without error. Therefore, the
# README.md file will be copied to the image every time even if there are no
# .crt files.
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

RUN apt-get update
RUN apt-get install -y --no-install-recommends jq openssh-server python3-pip python3-setuptools

RUN pip3 install --upgrade pip
RUN pip3 config set global.cert /etc/ssl/certs/ca-certificates.crt
RUN pip3 install requests python-dateutil redfishtool

COPY images/jump-host/assets/entrypoint.sh /
COPY images/jump-host/assets/sshd_config /etc/ssh

RUN useradd -m ubuntu

ENTRYPOINT ["/entrypoint.sh"]
