# AWS CloudFront with Signed URL
**This is the repository of [my blog post](https://minghsu0107.github.io/posts/aws-cloudfront-with-signed-url/)**.

This examples shows how to serve private contents on AWS S3 through CloudFront signed URL. We will be using [aws-sdk-go](https://github.com/aws/aws-sdk-go) as the programming client.
## Prerequisite
- A S3 bucket.
- A CloudFront distribution.
  - Should be created using the S3 owner because S3 bucket policies don’t apply to objects owned by other accounts.
- The CloudFront bucket access restriction is enabled.
- The CloudFront origin access identity is created and added to your S3 permission policy.
- The CloudFront viewer access restriction is enabled and associated with your key group.
- The public access of your S3 is blocked (default).
## Usage
```bash
S3_REGION=us-east-2 \
S3_ACCESS_KEY=my-s3-access-key \
S3_SECRET_KEY=my-s3-secret-key \
S3_BUCKET=my-s3-bucket \
CF_DOMAIN=mycfdomain.cloudfront.net \
CF_PUBLIC_KEY_ID=my-cloudfront-access-key \
CF_PRIKEY_PATH=my-cloudfront-prikey-path \
go run main.go
```
## Result
`hello.txt` will be uploaded to S3 bucket `my-s3-bucket` with key `mysubpath/hello.txt`. Its CloudFront URL `https://mycfdomain.cloudfront.net/mysubpath/hello.txt` will be signed, and the signed URL will be printed in the standard output. Users can access the object via this signed URL until it expires 1 hour later.
