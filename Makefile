-include .env

.PHONY: demo
demo:
	export Username=$(Username) &&export Password=$(Password) &&export Host=$(Host) &&\
	go run demo/consumer/main.go