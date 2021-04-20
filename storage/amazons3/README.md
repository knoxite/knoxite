# knoxite S3 storage backend

## URL Structure

The `amazons3` storage backend registers the `amazons3://` handler. To use it, supply a URL of the following format either as a `-r` parameter to your knoxite invocation or to the configuration system:

	amazons3://<bucket-name>/[prefix/][?region=REGION]&[endpoint=URL]&[force_path_style=true]

Optionally, some configuration parameters may be supplied as GET style parameters:

| Parameter Name | Valid values | Description |
| -------------- | ------------ | ----------- |
| `region` | valid AWS region descriptors | AWS Region. If this configuration is not specified, the backend falls back to other means of configuration, such as the `AWS_REGION` environment variable. |
| `endpoint` | valid URLs | **For testing purposes only**. This Parameter can be used to make S3 requests against backends other than those provided by AWS. This is not recommended.
| `force_path_style` | `true` | Use this parameter to force the underlying S3 SDK to make "path style" requests against the Amazon S3 backend. We don't recommend using this parameter if not required for compatibility reasons as [path style request are being sunset by AWS.][1]


## S3 bucket setup

For security reasons, we recommend setting up an S3 bucket and using short-lived credentials such as EC2 instance roles whenever feasible.

In situations where this is not possible, long-lived credentials such as IAM users are a good option.

In both cases, we recommend sticking to the principle of least privilege as outlined in [AWS IAM's best practices][2].

For inspiration, you can have a look at the [AWS Cloudformation template][3] we use to deploy testing infrastructure for this backend.

[1]: https://aws.amazon.com/blogs/aws/amazon-s3-path-deprecation-plan-the-rest-of-the-story/
[2]: https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html#grant-least-privilege
[3]: test_setup.template.yaml
