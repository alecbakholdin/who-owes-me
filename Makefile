dev: docker
	air
docker:
	docker compose -f docker-compose.test.yml up -d --force-recreate authelia_proxy actual_server actual_http_api
down:
	docker compose -f docker-compose.test.yml down
	docker compose down
