BASEDIR = $(shell pwd)
PROJECT=bingo-collab
SAACCOUNT=firebase-adminsdk-cffyr

env:
	gcloud config set project $(PROJECT)


dev:
	(trap 'kill 0' SIGINT; \
	cd $(BASEDIR)/backend && \
	export GOOGLE_APPLICATION_CREDENTIALS=$(BASEDIR)/creds/creds.json && \
	go run main.go firestore.go bingo.go & \
	cd $(BASEDIR)/frontend && ng serve --open )

	

server:
	cd $(BASEDIR)/backend && \
	go run main.go firestore.go bingo.go 

init:
	cd frontend && npm install
	cd backend && go mod vendor

serviceaccount:
	@echo ~~~~~~~~~~~~~ Download key for service account. 
	-gcloud iam service-accounts keys create creds/creds.json \
  	--iam-account $(SAACCOUNT)@$(PROJECT).iam.gserviceaccount.com  	