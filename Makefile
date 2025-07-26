VERSION=0.0.7

build:
	docker build -t cuongnb14/k8swatch:$(VERSION) .

push:
	docker push cuongnb14/k8swatch:$(VERSION)