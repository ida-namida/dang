# TODO: Handle if Heroku CLI is not installed
# TODO: Check if `heroku whoami` already returns an identity
heroku-login:
	heroku login -i
	heroku container:login

docker-build-tag-push:
	docker build --build-arg KOPURO_BASE_URL=$(KOPURO_BASE_URL) -f deploy/docker/dockerfile -t dang .
	docker tag dang:latest registry.heroku.com/$(HEROKU_APP_NAME)/web
	docker push registry.heroku.com/$(HEROKU_APP_NAME)/web

heroku-release:
	heroku container:release -a $(HEROKU_APP_NAME) web

heroku-docker-all: heroku-login docker-build-tag-push heroku-release