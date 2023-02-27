package apiclient_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var ctx context.Context

var _ = BeforeSuite(func() {
	ctx = context.Background()
})

func TestApiClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApiClient Suite")
}
