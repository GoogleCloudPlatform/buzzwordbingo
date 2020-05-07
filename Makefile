BASEDIR = $(shell pwd)


dev:
	(trap 'kill 0' SIGINT; \
	cd $(BASEDIR)/backend && \
	go run main.go firestore.go bingo.go & \
	cd $(BASEDIR)/frontend && ng serve --open )

	
