.PHONY: run
run:
	go run cmd/hr-helper/main.go

.PHONY: up
up:
	docker-compose up --build hr-helper

.PHONY: up-dev
up-dev:
	docker-compose up --build hr-helper-dev

.PHONY: deploy
deploy:
	./deploy.sh . dimasik@158.160.118.46:/home/dimasik/app ~/.ssh/yc_id_ed25519
