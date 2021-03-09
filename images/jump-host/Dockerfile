ARG BASE_IMAGE=gcr.io/google-appengine/python
FROM ${BASE_IMAGE}

# Inject custom root certificate authorities if needed.
# Docker does not have a good conditional copy statement and requires that a
# source file exists to complete the copy function without error. Therefore, the
# README.md file will be copied to the image every time even if there are no
# .crt files.
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

RUN apt-get update
RUN apt-get install -y --no-install-recommends jq

RUN pip3 config set global.cert /etc/ssl/certs/ca-certificates.crt
RUN pip3 install requests python-dateutil redfishtool

CMD ["/bin/bash"]
