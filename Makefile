VERSION = $(shell cat .version)

correct-pod-downafter:  main.go .version
	echo go build -ldflags "-X main.Version $(VERSION)"
	go build -ldflags "-X main.Version $(VERSION)"


clean:
	rm correct-pod-downafter


