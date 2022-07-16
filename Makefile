
run:
	go run .

build:
	go build -o build/main .

build-static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o build/main .
	# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/main .