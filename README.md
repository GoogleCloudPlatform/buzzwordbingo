# Bingomeeting

This project was generated with [Angular CLI](https://github.com/angular/angular-cli) version 9.0.6.

## Development server

Run `make dev` for a dev server. Navigate to `http://localhost:4200/`. The app will automatically reload if you change any of the js source files. If you update the Golang code, you will have to break out and rerun `make dev`


## Deploy to production

There are two ways to go about running this in production

1. Use `make deploy`
1. Setup a CLI pipeline on a git repo that runs a Cloud Build job to deploy to 
App Engine. There is a cloudbuild.yaml setup and a builder directory available 
for doing that. 

