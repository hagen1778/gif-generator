build:
	go build

run: build
	IMAGE_DIR=./testData COLLECTION_DIR=./ ./gif-generator