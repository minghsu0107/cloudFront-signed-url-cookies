# AWS CloudFront with Signed URL
This examples shows how to serve private contents on AWS S3 through CloudFront signed URL. We will be using [aws-sdk-go](https://github.com/aws/aws-sdk-go) as the programming client.
## Prerequisite
- A S3 bucket.
- A CloudFront distribution.
  - Should be created using the S3 owner because S3 bucket policies don’t apply to objects owned by other accounts.
- The CloudFront bucket access restriction is enabled.
- The CloudFront origin access identity is created and added to your S3 permission policy.
- The public access of your S3 is blocked (default).
- CloudFront Key ID and key pairs are created.
## Usage
```bash
S3_REGION=us-east-2 \
S3_ACCESS_KEY=my-s3-access-key \
S3_SECRET_KEY=my-s3-secret-key \
S3_BUCKET=my-s3-bucket \
CF_DOMAIN=mycfdomain.cloudfront.net \
CF_ACCESS_KEY=my-cloudfront-access-key \
CF_PRIKEY_PATH=my-cloudfront-prikey-path \
go run main.go
```
## Result
`hello.txt` will be uploaded to S3 bucket `my-s3-bucket` with key `mysubpath/hello.txt`. Its object URL `https://my-s3-bucket.s3.us-east-2.amazonaws.com/mysubpath/hello.txt` will be signed, and the signed URL will be printed in the standard output. Users can access the object via this signed URL until it expires 1 hour later.
