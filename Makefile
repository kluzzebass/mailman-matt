
run:
	go run .

tidy:
	go mod tidy

build:
	go build -o build/matt .

build-static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o build/matt .

docker-build:
	docker build -t kluzz/mailman-matt-go .

docker-run:
	docker run -p 3000:3000 -it --name mailman-matt --rm kluzz/mailman-matt-go
	
docker-run-detached:
	docker run -p 3000:3000 -d --name mailman-matt kluzz/mailman-matt-go

docker-rm:
	docker rm -f mailman-matt

docker-push:
	docker push kluzz/mailman-matt-go
