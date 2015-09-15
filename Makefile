test:
	ginkgo
	(cd conn && ginkgo)

.PHONY: test
