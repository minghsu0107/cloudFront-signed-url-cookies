package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

var (
	objKey         string = "mysubpath/hello.txt"
	objKeyWildcard string = "mysubpath/*"
	uploadFrom     string = "hello.txt"
)

var (
	s3Region     = os.Getenv("S3_REGION")
	accessKey    = os.Getenv("S3_ACCESS_KEY")
	secretKey    = os.Getenv("S3_SECRET_KEY")
	s3Bucket     = os.Getenv("S3_BUCKET")
	cfDomain     = os.Getenv("CF_DOMAIN")
	cfAccessKey  = os.Getenv("CF_PUBLIC_KEY_ID")
	cfPrikeyPath = os.Getenv("CF_PRIKEY_PATH")
)

func main() {
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	config := aws.Config{
		Credentials:      creds,
		Region:           s3Region,
		RetryMaxAttempts: 3,
	}
	s3Client := s3.NewFromConfig(config)

	fromFile, err := os.Open(uploadFrom)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", uploadFrom, err)
	}
	defer fromFile.Close()

	uploader := manager.NewUploader(s3Client)
	_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(objKey),
		Body:   fromFile,
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
	urlSigner := sign.NewURLSigner(cfAccessKey, priKey)

	rawURL := url.URL{
		Scheme: "https",
		Host:   cfDomain,
		Path:   objKey,
	}
	signedURL, err = urlSigner.Sign(rawURL.String(), time.Now().Add(1*time.Hour))
	if err != nil {
		exitErrorf("Failed to sign url, err: %v", err)
	}
	fmt.Printf("Get signed URL %q\n", signedURL)

	cookieSigner := sign.NewCookieSigner(cfAccessKey, priKey)
	policy := &sign.Policy{
		Statements: []sign.Statement{
			{
				Resource: "https://" + cfDomain + "/" + objKeyWildcard,
				Condition: sign.Condition{
					// Optional IP source address range
					// IPAddress: &sign.IPAddress{SourceIP: "192.0.2.0/24"},
					// Optional date URL is not valid until
					// DateGreaterThan: &sign.AWSEpochTime{time.Now().Add(30 * time.Minute)},
					// Required date the URL will expire after
					DateLessThan: &sign.AWSEpochTime{time.Now().Add(1 * time.Hour)},
				},
			},
		},
	}
	cookies, err := cookieSigner.SignWithPolicy(policy, func(o *sign.CookieOptions) {
		o.Path = "/"
		o.Domain = "localhost"
		o.Secure = false
	})
	if err != nil {
		exitErrorf("failed to create signed cookies, err: %v", err)
	}
	fmt.Printf("Created sign cookies: %v\n", cookies)

	content, err := httpGetWithCookie(rawURL.String(), cookies)
	if err != nil {
		exitErrorf("failed to access with signed cookies, err: %v", err)
	}
	fmt.Printf("Accessed with signed cookies, content: %v\n", content)

	router := gin.Default()
	router.GET("/auth", func(c *gin.Context) {
		for _, cookie := range cookies {
			http.SetCookie(c.Writer, cookie)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	router.Run(":80")
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func httpGetWithCookie(url string, cookies []*http.Cookie) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	for _, c := range cookies {
		// only add c.Name and c.Value
		req.AddCookie(c)
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)

	return string(data), nil
}
