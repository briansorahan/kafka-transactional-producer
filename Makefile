DOCKER_NETWORK = kafka-transactional-producer_default
IMG            = transactional-producer-test

kafka:
	@docker-compose -f $@.yml up

clean-kafka:
	@docker-compose -f $(subst clean-,,$@).yml rm --force --stop

# maybe the prune commands are a bit of overkill?
clean: clean-kafka
	@docker system prune -f
	@docker volume prune -f

topics:
	@docker-compose -f $@.yml up

.image: Dockerfile $(wildcard *.go)
	@docker build -t $(IMG) .
	@touch $@

run: .image
	@docker run --rm --network $(DOCKER_NETWORK) $(IMG)

.PHONY: clean clean-kafka kafka run
