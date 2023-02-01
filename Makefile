.PHONY:
	install-tools

# Install build tools and other required software.
install-tools:
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	go install github.com/onsi/gomega
