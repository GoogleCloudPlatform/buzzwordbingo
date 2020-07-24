BASEDIR = $(shell pwd)
PROJECT=$(BINGO_PROJECT_ID)
REDISNAME=bingocollab
REGION=us-central1
GAEREGION=us-central
SAACCOUNT=bingo-developer-account
PROJECTNUMBER=$(shell gcloud projects list --filter="$(PROJECT)" \
			--format="value(PROJECT_NUMBER)")
REDISIP=$(shell gcloud beta redis instances describe $(REDISNAME) \
			--region $(REGION) --format='value(host)')
VPCCONNECTOR=$(shell gcloud compute networks vpc-access connectors describe \
			$(REDISNAME)connector --region $(REGION) --format='value(name)' )
env:
	gcloud config set project $(PROJECT)

clean:
	-rm -rf backend/static		

frontend: clean
	cd frontend && ng build --prod

deploy: env frontend
	cd backend && gcloud app deploy -q

build:
	gcloud builds submit --config cloudbuild.yaml --timeout=1200s \
	--machine-type=n1-highcpu-8 . 	

init:
	cd frontend && npm install
	cd backend && go mod vendor

project: env services appengine cloudbuild memorystore \
		serviceaccount perms firestore-rules 	

services: env
	-gcloud services enable vpcaccess.googleapis.com
	-gcloud services enable cloudbuild.googleapis.com
	-gcloud services enable appengine.googleapis.com 
	-gcloud services enable redis.googleapis.com
	-gcloud services enable firestore.googleapis.com
	-gcloud services enable iap.googleapis.com		

appengine: env
	@echo ~~~ Intialize AppEngine on $(PROJECT)
	-gcloud app create --region $(GAEREGION)	

cloudbuild: env
	@echo ~~~ Enable Build service account to deploy to GAE on $(PROJECT)
	-gcloud projects add-iam-policy-binding $(PROJECT) \
  	--member serviceAccount:$(PROJECTNUMBER)@cloudbuild.gserviceaccount.com \
  	--role roles/appengine.appAdmin	

memorystore: env
	-gcloud redis instances create $(REDISNAME) --size=1 --region=$(REGION)
	-gcloud compute networks vpc-access connectors create \
	$(REDISNAME)connector --network default --region $(REGION) \
	--range 10.8.0.0/28 	

serviceaccount: env
	@echo ~~~ Create service account for Development   
	-gcloud iam service-accounts create $(SAACCOUNT) \
    --description "A service account for development of a bingo game" \
    --display-name "Bingo App" --project $(PROJECT)
	@echo ~~~ Download key for service account. 
	-gcloud iam service-accounts keys create creds/creds.json \
  	--iam-account $(SAACCOUNT)@$(PROJECT).iam.gserviceaccount.com  	

perms:
	@echo ~~~ Grant Service account permissions
	-gcloud projects add-iam-policy-binding $(PROJECT) \
  	--member serviceAccount:$(PROJECT)@appspot.gserviceaccount.com \
  	--role roles/vpaccess.user
	-gcloud projects add-iam-policy-binding $(PROJECT) \
  	--member serviceAccount:$(PROJECTNUMBER)@cloudbuild.gserviceaccount.com \
  	--role roles/vpaccess.user
	-gcloud projects add-iam-policy-binding $(PROJECT) \
  	--member serviceAccount:$(PROJECTNUMBER)@cloudbuild.gserviceaccount.com \
  	--role roles/vpaccess.user  
	-gcloud projects add-iam-policy-binding $(PROJECT) \
  	--member serviceAccount:$(SAACCOUNT)@$(PROJECT).iam.gserviceaccount.com \
  	--role roles/compute.networkAdmin
	-gcloud projects add-iam-policy-binding $(PROJECT) \
  	--member serviceAccount:$(SAACCOUNT)@$(PROJECT).iam.gserviceaccount.com \
  	--role roles/project.viewer  

firestore-rules:
	firebase deploy --only firestore


secure: env
	gcloud services enable cloudresourcemanager.googleapis.com
	gcloud services enable iap.googleapis.com
	gcloud iap web enable --resource-type=app-engine \
	--oauth2-client-id $(BINGO_OAUTH_ID) \
	--oauth2-client-secret $(BINGO_OAUTH_SECRET)
	gcloud iap web add-iam-policy-binding  \
      --member='allAuthenticatedUsers' \
      --role='roles/iap.httpsResourceAccessor' 	




listvpc:
	echo $(VPCCONNECTOR)



redis: redisclean
	docker run --name some-redis -p 6379:6379 -d redis	

redisclean:
	-docker stop some-redis
	-docker rm some-redis

dev: redis
	(trap 'kill 0' SIGINT; \
	cd $(BASEDIR)/backend && \
	export REDISHOST=127.0.0.1 && \
	export REDISPORT=6379 && \
	export GOOGLE_APPLICATION_CREDENTIALS=$(BASEDIR)/creds/creds.json && \
	go run main.go firestore.go bingo.go cache.go game.go & \
	cd $(BASEDIR)/frontend && ng serve --open )		

server:  
	cd $(BASEDIR)/backend && \
	export REDISHOST=127.0.0.1 && \
	export REDISPORT=6379 && \
	export GOOGLE_APPLICATION_CREDENTIALS=$(BASEDIR)/creds/creds.json && \
	go run main.go firestore.go bingo.go cache.go game.go

fe: 
	cd $(BASEDIR)/frontend && ng serve --open



savecreds: env
	-gsutil mb gs://$(PROJECT)_creds/
	-gsutil cp $(BASEDIR)/frontend/src/environments/environment.ts gs://$(PROJECT)_creds/environment.ts
	-gsutil cp $(BASEDIR)/frontend/src/environments/environment.prod.ts gs://$(PROJECT)_creds/environment.prod.ts
	-gsutil cp $(BASEDIR)/backend/app.yaml gs://$(PROJECT)_creds/app.yaml

builders:
	cd builders/gotester && make build
	cd builders/ng && make build	

function: env
	gcloud services enable cloudfunctions.googleapis.com
	-gcloud functions deploy subscribeMailgun --trigger-topic cloud-builds \
	--runtime nodejs10 --set-env-vars GCLOUD_PROJECT=$(PROJECT) \
	--source $(BASEDIR)/functions/email	--allow-unauthenticated	
	-gcloud alpha functions add-iam-policy-binding subscribeMailgun \
	--member=allUsers --role=roles/cloudfunctions.invoker

schedule: env
	-gcloud services enable cloudscheduler.googleapis.com
	gcloud scheduler jobs create app-engine purgeGames --schedule="0 1 * * *"	\
	--relative-url="/api/game/purge"


fsconfig:
	firebase firestore:indexes