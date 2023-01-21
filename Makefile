ifneq ("$(wildcard .env)", "")
	include .env
endif

build:
	go build

run: build
	./bitheroes-community-bot

clean:
	rm bitheroes-community-bot

docker_build:
	docker build -t $(IMAGE_URL) .

docker_run: docker_build
	docker run -e "BITHEROES_COMMUNITY_BOT_AUTH_TOKEN=$(BITHEROES_COMMUNITY_BOT_AUTH_TOKEN)" $(IMAGE_URL)

docker_push: docker_build
	docker push $(IMAGE_URL)

k8s_deploy:
	kubectl apply -f k8s
