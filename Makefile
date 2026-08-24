.PHONY: catalog-info.yaml
catalog-info.yaml:
	cd scripts/generate-catalog-info && \
	go run . -root-dir ../../ -output-path ../../catalog-info.yaml
