ARG BASE_IMAGE=gcr.io/google-appengine/python
FROM ${BASE_IMAGE}

RUN apt-get update
RUN apt-get install -y --no-install-recommends jq

RUN pip3 install requests python-dateutil redfishtool

CMD ["/bin/bash"]
