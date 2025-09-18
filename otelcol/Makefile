.PHONY: build
build:
	ocb --skip-compilation --config builder-config.yaml

.PHONY: docker-image
docker-image: build
	docker build . -t otelcol:0.0.1

.PHONY: clean
clean:
	rm -rf build
