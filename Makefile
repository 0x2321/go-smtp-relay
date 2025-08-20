build_amd:
	GOOS=linux GOARCH=amd64 go build -o smtp-relay

build_arm:
	GOOS=linux GOARCH=arm64 go build -o smtp-relay