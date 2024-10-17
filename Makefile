.PHONY: test

test:
	docker compose up -d --wait postgres-test mysql-test
	go test -short -count=1 -coverpkg=./handler,./service,./repository ./handler
