
run:
	go run ./main.go

test:
	go test ./...

PAYLOAD := $(shell echo '{"action": "INSERT", "tag": "sakajunquality/flow:abc123"}' | base64)
test-message:
	curl -X POST http://localhost:8080 \
	-H "Authorization: Bearer $$(gcloud auth print-identity-token)" \
	-H 'Content-Type: application/json' \
	-d '{"message": {"data": "$(PAYLOAD)"}}'
