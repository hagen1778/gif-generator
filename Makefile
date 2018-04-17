build:
	go build

run: build
	IMAGE_DIR=./testData PATH_DATASET=./ PATH_MODEL=./README.md ./gif-generator