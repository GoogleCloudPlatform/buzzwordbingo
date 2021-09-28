BASEDIR = $(shell pwd)

frontend: clean
	cd ../frontend && ng build --prod --output-path=../container/static

backend:
	cp ../backend/bingo.go .
	cp ../backend/cache.go .
	cp ../backend/firestore.go .
	cp ../backend/game.go .
	cp ../backend/main.go .
	cp ../backend/go.mod .
	cp ../backend/go.sum .

redis: redisclean
	docker run --name redis -p 6379:6379 -d redis	

redisclean:
	-docker stop some-redis
	-docker rm some-redis

build:
	docker build --tag bingo-services .     

run: redis
	docker run --publish 8080:8080 \
	-e GOOGLE_APPLICATION_CREDENTIALS=/tmp/keys/creds.json \
	-v $(BASEDIR)/../creds/creds.json:/tmp/keys/creds.json:ro  \
	-e REDISHOST=redis -e REDISPORT=6379 --link redis:redis  \
	bingo-services 

clean:
	-rm *.go	
	-rm go.*	
	-rm -rf static	