build:
	go build

run: build
	PATH_RESULT=./testData PATH_DATASET=./ ./gif-generator