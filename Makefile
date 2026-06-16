dev: docker
	air
docker:
	docker compose -f docker-compose.test.yml up -d authelia actual_server actual_http_api
