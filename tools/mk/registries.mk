.PHONY: gen-db gen-check

gen-db:
	go run ./tools/genmanifests db-gen

gen-check:
	go run ./tools/genmanifests db-check
