# AWS CloudFront with Signed URL
**This is the repository of [my blog post](https://minghsu0107.github.io/posts/aws-cloudfront-with-signed-url/)**.

This example shows how to serve private contents on AWS S3 through CloudFront signed URL and signed cookies. We will be using [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) as the programming client.
## Prerequisite
- A S3 bucket.
- A CloudFront distribution.
  - Should be created using the S3 owner because S3 bucket policies don’t apply to objects owned by other accounts.
- The CloudFront bucket access restriction is enabled.
- A CloudFront origin access identity is created and added to your S3 permission policy.
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
1. `hello.txt` will be uploaded to S3 bucket `my-s3-bucket` with key `mysubpath/hello.txt`. Its CloudFront URL `https://mycfdomain.cloudfront.net/mysubpath/hello.txt` will be signed, and the signed URL will be printed to standard output. Users can access the object via this signed URL until it expires after 1 hour.
2. Signed cookies will be returned and printed to standard output. The signed cookies use the following custom policy:
    - Allow users to access `https://mycfdomain.cloudfront.net/mysubpath/*` (wildcard).
    - Signed cookies will expire after 1 hour.
3. The program will request `https://mycfdomain.cloudfront.net/mysubpath/hello.txt` with signed cookies and print the content of `hello.txt` to standard output.
4. An http server will be started. Users can set signed cookies via `GET http://localhost/auth`. The following cookies will be set: `CloudFront-Signature`, `CloudFront-Policy`, and `CloudFront-Key-Pair-Id`.
