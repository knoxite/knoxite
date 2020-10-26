/*
 * knoxite
 *     Copyright (c) 2020, Johannes Fürmann <fuermannj+floss@gmail.com>
 *
 *   For license see LICENSE
 */

package amazons3

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/knoxite/knoxite"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type mockS3Client struct {
	s3iface.S3API
	getObjectOutput    *s3.GetObjectOutput
	getObjectError     error
	deleteObjectOutput *s3.DeleteObjectOutput
	deleteObjectError  error
	putObjectOutput    *s3.PutObjectOutput
	putObjectError     error
	headObjectOutput   *s3.HeadObjectOutput
	headObjectError    error
}

func (mc *mockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return mc.getObjectOutput, mc.getObjectError
}

func (mc *mockS3Client) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return mc.deleteObjectOutput, mc.deleteObjectError
}

func (mc *mockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return mc.putObjectOutput, mc.putObjectError
}

func (mc *mockS3Client) HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	return mc.headObjectOutput, mc.headObjectError
}

var _ = Describe("Stat", func() {
	var (
		backend    knoxite.BackendFilesystem
		err        error
		sizeActual uint64
	)

	sizeExpected := uint64(23)

	When("The object exists", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{
					headObjectOutput: &s3.HeadObjectOutput{
						ContentLength: aws.Int64(int64(sizeExpected)),
					},
				},
			}

			sizeActual, err = backend.Stat("asdf")
		})

		It("Should return the correct size", func() {
			Expect(sizeActual).To(Equal(sizeExpected))
		})

		It("shouldn't return an error", func() {
			Expect(err).To(BeNil())
		})
	})

	When("The function doesn't exist", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{
					headObjectError: awserr.New("NotFound", "lel", fmt.Errorf("lel")),
				},
			}

			sizeActual, err = backend.Stat("asdf")
		})

		It("should return a size of 0", func() {
			Expect(sizeActual).To(BeZero())
		})

		It("should return an error", func() {
			Expect(err).NotTo(BeNil())
		})
	})
})

var _ = Describe("ReadFile", func() {
	var (
		backend knoxite.BackendFilesystem
		err     error
		result  []byte
	)

	file := []byte("hellohello")

	When("The file was read correctly", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{
					getObjectOutput: &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBuffer(file)),
					},
				},
			}

			result, err = backend.ReadFile("asdf")
		})

		It("doesn't return an error", func() {
			Expect(err).To(BeNil())
		})

		It("returns the file's content", func() {
			Expect(result).To(Equal(file))
		})
	})

	When("an error occurred in the backend", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{
					getObjectError: awserr.New("NoSuchBucket", "lel", fmt.Errorf("lel")),
				},
			}

			result, err = backend.ReadFile("asdf")
		})

		It("returns an empty byte array", func() {
			Expect(result).To(BeEmpty())
		})

		It("returns an error message", func() {
			Expect(err).ToNot(BeNil())
		})

	})
})

var _ = Describe("WriteFile", func() {
	var (
		backend knoxite.BackendFilesystem
		err     error
		size    uint64
	)

	file := []byte("asdfasdf")

	When("the file was written correctly", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{},
			}

			size, err = backend.WriteFile("asdf", file)
		})

		It("should return the file's size", func() {
			Expect(size).To(Equal(uint64(len(file))))
		})

		It("shouldn't return an error", func() {
			Expect(err).To(BeNil())
		})
	})

	When("there was an error writing the file", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{
					putObjectError: awserr.New("NoSuchBucket", "lol", fmt.Errorf("lel")),
				},
			}

			size, err = backend.WriteFile("asdf", file)
		})

		It("should return a file size of zero", func() {
			Expect(size).To(BeZero())
		})

		It("should return an error", func() {
			Expect(err).ToNot(BeNil())
		})
	})
})

var _ = Describe("DeleteFile", func() {
	var (
		backend knoxite.BackendFilesystem
		err     error
	)

	When("the path was deleted successfully", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{},
			}
			err = backend.DeleteFile("asdf")
		})

		It("shouldn't return an error", func() {
			Expect(err).To(BeNil())
		})
	})

	When("there was an error deleting the path", func() {
		BeforeEach(func() {
			backend = &AmazonS3StorageBackend{
				service: &mockS3Client{
					deleteObjectError: awserr.New("NotFound", "Foobar", fmt.Errorf("NotFound")),
				},
			}
			err = backend.DeleteFile("asdf")
		})

		It("should return an error", func() {
			Expect(err).ToNot(BeNil())
		})
	})

})

var _ = Describe("CreatePath", func() {
	var (
		backend knoxite.BackendFilesystem
		err     error
	)

	BeforeEach(func() {
		backend = &AmazonS3StorageBackend{}
		err = backend.CreatePath("foo")
	})

	It("should return `nil`", func() {
		Expect(err).To(BeNil())
	})
})

var _ = Describe("Close", func() {
	var (
		backend knoxite.Backend
		err     error
	)

	BeforeEach(func() {
		url, _ := url.Parse("amazons3://foobarfoo/asdgf")
		backend, err = (&AmazonS3StorageBackend{}).NewBackend(*url)
		err = backend.Close()
	})

	It("should return `nil`", func() {
		Expect(err).To(BeNil())
	})
})

var _ = Describe("AvailableSpace", func() {
	var (
		backend knoxite.Backend
		space   uint64
		err     error
	)

	BeforeEach(func() {
		url, _ := url.Parse("amazons3://foobarfoo/asdgf")
		backend, err = (&AmazonS3StorageBackend{}).NewBackend(*url)
		space, err = backend.AvailableSpace()
	})

	It("should return `ErrAvailableSpaceUnlimited`", func() {
		Expect(err).To(Equal(knoxite.ErrAvailableSpaceUnlimited))
	})

	It("should return 0 as available space", func() {
		Expect(space).To(Equal(uint64(0)))
	})
})
