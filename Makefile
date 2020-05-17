BASEDIR = $(shell pwd)
PROJECT=bingo-collab
SAACCOUNT=firebase-adminsdk-cffyr
PROJECTNUMBER=$(shell gcloud projects list --filter="$(PROJECT)" --format="value(PROJECT_NUMBER)")

env:
	gcloud config set project $(PROJECT)


clean:
	-rm -rf server/dist		

frontend: clean
	cd frontend && ng build --prod


deploy: env clean frontend
	cd backend && gcloud app deploy -q

dev: fixvendor
	(trap 'kill 0' SIGINT; \
	cd $(BASEDIR)/backend && \
	export GOOGLE_APPLICATION_CREDENTIALS=$(BASEDIR)/creds/creds.json && \
	go run main.go firestore.go bingo.go & \
	cd $(BASEDIR)/frontend && ng serve --open )

fixvendor:
	@echo Copying fix library not working. 
	cp $(BASEDIR)/vendorfix/validate.go	$(BASEDIR)/backend/vendor/google.golang.org/api/idtoken/validate.go

server: fixvendor
	cd $(BASEDIR)/backend && \
	export GOOGLE_APPLICATION_CREDENTIALS=$(BASEDIR)/creds/creds.json && \
	go run main.go firestore.go bingo.go 

init:
	cd frontend && npm install
	cd backend && go mod vendor

serviceaccount: env
	@echo ~~~~~~~~~~~~~ Download key for service account. 
	-gcloud iam service-accounts keys create creds/creds.json \
  	--iam-account $(SAACCOUNT)@$(PROJECT).iam.gserviceaccount.com  	

project: env
	@echo ~~~~~~~~~~~~~ Intialize AppEngine on $(PROJECT)
	-gcloud app create --region us-central -q 
	@echo ~~~~~~~~~~~~~ Enable Cloud Build service account to deploy to AppEngine on $(PROJECT)
	-gcloud projects add-iam-policy-binding $(PROJECT) \
  	--member serviceAccount:$(PROJECTNUMBER)@cloudbuild.gserviceaccount.com \
  	--role roles/appengine.appAdmin	
	@echo ~~~~~~~~~~~~~ Create Angular builder for Cloud Build 
	-cd builder && make build    