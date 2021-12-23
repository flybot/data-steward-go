cert:
	cd cert; ./gen.sh; cd ..

test:
	cd web-server; go test -cover -race ./...; cd ..

.PHONY: cert test
