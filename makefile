default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./account/. -v $(TESTARGS) -timeout 120m
