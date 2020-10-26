/*
 * knoxite
 *     Copyright (c) 2020, Johannes Fürmann <fuermannj+floss@gmail.com>
 *
 *   For license see LICENSE
 */

package amazons3

import (
	"net/url"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestBDD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "amazons3 suite")
}

var _ = Describe("Factory", func() {
	DescribeTable(
		"Instantiating a backend",
		func(inputUrl string, expectedErr error) {
			parsedURL, err := url.Parse(inputUrl)
			if err != nil {
				panic("input url invalid")
			}
			factory := &AmazonS3StorageBackend{}
			_, err = factory.NewBackend(*parsedURL)
			if expectedErr == nil {
				Expect(err).To(BeNil())
				return
			}
			Expect(err).To(Equal(expectedErr))
		},
		Entry(
			"simple url",
			"amazons3://asdfbucket/foobar",
			nil,
		),
		Entry(
			"url with region",
			"amazons3://asdfbucket/foobar?region=eu-west-1",
			nil,
		),
		Entry(
			"url with endpoint",
			"amazons3://asdfbucket/foobar?endpoint=http://localhost:1337",
			nil,
		),
	)
})
