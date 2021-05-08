package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	objKey     string = "mysubpath/hello.txt"
	uploadFrom string = "hello.txt"
)

var (
	s3Region     = os.Getenv("S3_REGION")
	accessKey    = os.Getenv("S3_ACCESS_KEY")
	secretKey    = os.Getenv("S3_SECRET_KEY")
	s3Bucket     = os.Getenv("S3_BUCKET")
	cfDomain     = os.Getenv("CF_DOMAIN")
	cfAccessKey  = os.Getenv("CF_ACCESS_KEY")
	cfPrikeyPath = os.Getenv("CF_PRIKEY_PATH")
)

func main() {
	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")

	config := &aws.Config{
		Credentials: creds,
		Region:      aws.String(s3Region),
		MaxRetries:  aws.Int(3),
	}
	session, err := session.NewSession(config)
	if err != nil {
		fmt.Println("failed to create session", err)
		return
	}

	fromFile, err := os.Open(uploadFrom)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", uploadFrom, err)
	}
	defer fromFile.Close()

	uploader := s3manager.NewUploader(session)
	var output *s3manager.UploadOutput
	output, err = uploader.UploadWithContext(context.Background(), &s3manager.UploadInput{
		Bucket:  aws.String(s3Bucket),
		Key:     aws.String(objKey),
		Body:    fromFile,
		Expires: aws.Time(time.Now().Local().Add(3 * time.Hour)),
	})
	if err != nil {
		// Print the error and exit.
		exitErrorf("Unable to upload %q to %q, %v", uploadFrom, s3Bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", uploadFrom, s3Bucket)

	var priKeyFile *os.File
	priKeyFile, err = os.Open(cfPrikeyPath)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", cfPrikeyPath, err)
	}

	var priKey *rsa.PrivateKey
	priKey, err = sign.LoadPEMPrivKey(priKeyFile)
	if err != nil {
		exitErrorf("err loading private key, %v", err)
	}

	var signedURL string
	signer := sign.NewURLSigner(cfAccessKey, priKey)
	signedURL, err = signer.Sign(output.Location, time.Now().Add(1*time.Hour))
	if err != nil {
		exitErrorf("Failed to sign url, err: %v", err)
	}

	u, err := url.Parse(signedURL)
	if err != nil {
		exitErrorf("Failed to parse signed url %v, err: %v", signedURL, err)
	}
	u.Host = cfDomain
	signedURL = u.String()
	fmt.Printf("Get signed URL %q\n", signedURL)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
