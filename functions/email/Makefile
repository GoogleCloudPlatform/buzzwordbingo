BASEDIR = $(shell pwd)
APPNAME=color
ZONE=us-central1-a
REGION=us-central1
PORT=:80

env:
	@echo "Making sure project settings all in order"
	@gcloud config set project $(STUBHUB_PROJECT)
	@gcloud config set compute/zone $(ZONE)
	@gcloud config set run/region $(ZONE)


deploy: env
	-gcloud functions deploy subscribeMailgun --trigger-topic cloud-builds --runtime nodejs10