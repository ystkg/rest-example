.PHONY: test testshort testall testdbup

test: testdbup
	go test -short -count=1 -coverpkg=./handler,./service,./repository ./handler

testshort:
	go test -short -count=1 -coverpkg=./handler,./service,./repository ./handler

testall: testdbup
	go test -count=1 -coverpkg=./handler,./service,./repository ./handler

testdbup:
	docker compose up -d --wait postgres-test mysql-test
