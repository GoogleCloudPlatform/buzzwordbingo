FROM golang:stretch
RUN apt-get update && apt-get -y upgrade && apt-get install wget default-jre default-jdk  -y 
RUN curl https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz > /tmp/google-cloud-sdk.tar.gz
RUN mkdir -p /usr/local/gcloud \
  && tar -C /usr/local/gcloud -xvf /tmp/google-cloud-sdk.tar.gz \
  && /usr/local/gcloud/google-cloud-sdk/install.sh
ENV PATH $PATH:/usr/local/gcloud/google-cloud-sdk/bin
RUN gcloud components install beta 
RUN gcloud components install cloud-firestore-emulator
RUN chmod +x /usr/local/go/bin/go
