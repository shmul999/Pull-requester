DOCKER_COMPOSE_FILE=build/docker-compose.yml

up:
	cd build && docker-compose up --build

down:
	cd build && docker-compose down

restart: down up

logs:
	cd build && docker-compose logs -f

ps:
	cd build && docker-compose ps

clean:
	cd build && docker-compose down -v
	docker system prune -f

rebuild:
	cd build && docker-compose down
	cd build && docker-compose up --build
