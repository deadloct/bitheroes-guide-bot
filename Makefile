ifneq ("$(wildcard .env)", "")
	include .env
endif

NAME := bitheroes-guide-bot
LOCAL_PATH := bin

clean:
	rm -rf $(LOCAL_PATH)

build: clean
	go build -o $(LOCAL_PATH)/$(NAME)
	cp -R data bin/

run: build
	$(LOCAL_PATH)

ssh_build: clean
	GOOS=linux GOARCH=arm GOARM=5 go build -o $(LOCAL_PATH)/$(NAME)
	cp -R data bin/

ssh_deploy: ssh_build
	rsync -avz bin/ $(SSH_HOST):$(SSH_DIR)
	-ssh $(SSH_HOST) "killall $(NAME)"

docker_build:
	docker build -t $(IMAGE_URL) .

docker_run: docker_build
	docker run -e "BITHEROES_GUIDE_BOT_AUTH_TOKEN=$(BITHEROES_GUIDE_BOT_AUTH_TOKEN)" $(IMAGE_URL)

docker_push: docker_build
	docker push $(IMAGE_URL)

k8s_deploy:
	kubectl apply -f k8s
