package cos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/limingyao/excellent-go/config"
	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Host       string `yaml:"host"`
	BucketName string `yaml:"bucket_name"`
	AppId      string `yaml:"app_id"`
	Region     string `yaml:"region"`
	SecretId   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	Token      string `yaml:"token"`
}

func (c *Configuration) Init() error {
	return nil
}

var (
	spaceRegexp = regexp.MustCompile(`[ \n\r\t\v\f]+`)

	// Virtual-Hosted-Style:
	//   http://BUCKET.s3.amazonaws.com/KEY
	//   https://bucket.s3.Region.amazonaws.com/key
	//   https://bucket.cos.Region.myqcloud.com/key
	virtualHostedStyle = regexp.MustCompile(`^(?:https?://)?([\w-]+)-([0-9]+)(?:[0-9a-zA-Z:.-]+)?/([^?]*)(?:\?(.*))?$`)

	// Path-Style:
	//   http://s3.amazonaws.com/BUCKET/KEY
	//   https://s3.Region.amazonaws.com/bucket/key
	//   https://cos.Region.myqcloud.com/bucket/key
	pathStyle = regexp.MustCompile(`^(?:https?://)?(?:[0-9a-zA-Z:.-]+/)?([\w-]+)-([0-9]+)/([^?]*)(?:\?(.*))?$`)
)

func ParseVirtualHostedStyle(url string) (bucketName string, appId string, key string, err error) {
	matched := virtualHostedStyle.FindStringSubmatch(url)
	if len(matched) != 5 {
		return "", "", "", fmt.Errorf("parse virtual-hosted-style [%s] failed", url)
	}
	return matched[1], matched[2], matched[3], nil
}

func ParsePathStyle(url string) (bucketName string, appId string, key string, err error) {
	matched := pathStyle.FindStringSubmatch(url)
	if len(matched) != 5 {
		return "", "", "", fmt.Errorf("parse path-style [%s] failed", url)
	}
	return matched[1], matched[2], matched[3], nil
}

func formatError(err error) error {
	if err == nil {
		return nil
	}
	var awsErr awserr.Error
	if ok := errors.As(err, &awsErr); ok {
		var awsReqErr awserr.RequestFailure
		if ok := errors.As(err, &awsReqErr); ok {
			return fmt.Errorf("{status_code: %d, request_id: %s, code: %s, msg: %s, orig_err: %v}",
				awsReqErr.StatusCode(), awsReqErr.RequestID(), awsErr.Code(), awsErr.Message(), formatError(awsErr.OrigErr()))
		}
		return fmt.Errorf("{code: %s, msg: %s, orig_err: %v}", awsErr.Code(), awsErr.Message(), formatError(awsErr.OrigErr()))
	}
	msg := err.Error()
	msg = spaceRegexp.ReplaceAllString(msg, " ")
	return fmt.Errorf("%s", msg)
}

type Bucket struct {
	client *s3.S3

	bucketName string
	appId      string

	pathStyle bool
}

type Endpoint func(hostPort, region string) string

func QCloudEndpoint(host, region string) string {
	return fmt.Sprintf("https://cos.%s.myqcloud.com", region)
}

func RawEndpoint(host, region string) string {
	return host
}

func NewBucket(c *Configuration, endpoint Endpoint, opts ...Option) (*Bucket, error) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	// transport
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = defaultOpts.maxIdleConns
	transport.MaxIdleConnsPerHost = defaultOpts.maxIdleConns

	// cfg
	cfg := &aws.Config{
		Region:           aws.String(c.Region),
		Endpoint:         aws.String(endpoint(c.Host, c.Region)),
		S3ForcePathStyle: aws.Bool(defaultOpts.pathStyle),
		Credentials:      credentials.NewStaticCredentials(c.SecretId, c.SecretKey, c.Token),
		DisableSSL:       aws.Bool(defaultOpts.disableSSL),
		HTTPClient:       &http.Client{Transport: transport},
	}

	// session
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	bucket := &Bucket{}
	bucket.client = s3.New(sess)
	bucket.bucketName = c.BucketName
	bucket.appId = c.AppId
	bucket.pathStyle = defaultOpts.pathStyle
	return bucket, nil
}

func (x Bucket) parse(url string) (bucketName string, appId string, key string, err error) {
	if x.pathStyle {
		return ParsePathStyle(url)
	}
	return ParseVirtualHostedStyle(url)
}

func (x Bucket) Client() *s3.S3 {
	return x.client
}

func (x Bucket) Put(ctx context.Context, key string, buffer []byte) (string, error) {
	bufferReader := bytes.NewReader(buffer)
	key = strings.TrimPrefix(key, "/")

	if int64(len(buffer)) >= s3manager.DefaultUploadPartSize {
		uploader := s3manager.NewUploaderWithClient(x.client)
		response, err := uploader.UploadWithContext(ctx,
			&s3manager.UploadInput{
				Body:   bufferReader,
				Bucket: aws.String(fmt.Sprintf("%s-%s", x.bucketName, x.appId)),
				Key:    aws.String(key),
			},
			func(u *s3manager.Uploader) {
				u.PartSize = s3manager.DefaultUploadPartSize
				u.Concurrency = s3manager.DefaultUploadConcurrency
				u.LeavePartsOnError = true
			},
		)
		if err != nil {
			return "", formatError(err)
		}
		return response.Location, nil
	}

	_, err := x.client.PutObjectWithContext(ctx,
		&s3.PutObjectInput{
			Body:   aws.ReadSeekCloser(bufferReader),
			Bucket: aws.String(fmt.Sprintf("%s-%s", x.bucketName, x.appId)),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return "", formatError(err)
	}

	// 获取地址
	response, _ := x.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("%s-%s", x.bucketName, x.appId)),
		Key:    aws.String(key),
	})
	getUrl, err := response.Presign(time.Minute)
	if err != nil {
		return "", formatError(err)
	}
	t, err := url.Parse(getUrl)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s%s", t.Scheme, t.Host, t.Path), nil
}

func (x Bucket) Download(ctx context.Context, url string) ([]byte, error) {
	bucketName, appId, sourceKey, err := x.parse(url)
	if err != nil {
		return nil, err
	}

	bufferWriter := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloaderWithClient(x.client)
	if _, err = downloader.DownloadWithContext(ctx, bufferWriter, &s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("%s-%s", bucketName, appId)),
		Key:    aws.String(sourceKey),
	}); err != nil {
		return nil, formatError(err)
	}
	return bufferWriter.Bytes(), nil
}

func (x Bucket) Get(ctx context.Context, url string) ([]byte, error) {
	bucketName, appId, sourceKey, err := x.parse(url)
	if err != nil {
		return nil, err
	}

	obj, err := x.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("%s-%s", bucketName, appId)),
		Key:    aws.String(sourceKey),
	})
	if err != nil {
		return nil, formatError(err)
	}
	defer func() {
		if err := obj.Body.Close(); err != nil {
			log.WithError(err).Warn()
		}
	}()

	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (x Bucket) Copy(ctx context.Context, sourceUrl, targetKey string) (string, error) {
	targetKey = strings.TrimPrefix(targetKey, "/")
	bucketName, appId, sourceKey, err := x.parse(sourceUrl)
	if err != nil {
		return "", err
	}

	if _, err = x.client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(fmt.Sprintf("%s-%s", x.bucketName, x.appId)),
		CopySource: aws.String(url.PathEscape(fmt.Sprintf("%s-%s/%s", bucketName, appId, sourceKey))),
		Key:        aws.String(targetKey),
	}); err != nil {
		return "", formatError(err)
	}

	// 获取地址
	response, _ := x.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("%s-%s", x.bucketName, x.appId)),
		Key:    aws.String(targetKey),
	})
	getUrl, err := response.Presign(time.Minute)
	if err != nil {
		return "", formatError(err)
	}
	t, err := url.Parse(getUrl)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s%s", t.Scheme, t.Host, t.Path), nil
}

func (x Bucket) Delete(ctx context.Context, url string) error {
	bucketName, appId, sourceKey, err := x.parse(url)
	if err != nil {
		return err
	}

	if _, err = x.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(fmt.Sprintf("%s-%s", bucketName, appId)),
		Key:    aws.String(sourceKey),
	}); err != nil {
		return formatError(err)
	}
	return nil
}

func (x Bucket) List(ctx context.Context, url string, latestKey *string, maxKeys int64) ([]string, *string, error) {
	bucketName, appId, sourceKey, err := x.parse(url)
	if err != nil {
		return nil, nil, err
	}

	files := make([]string, 0)
	objs, err := x.client.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket:    aws.String(fmt.Sprintf("%s-%s", bucketName, appId)),
		Prefix:    aws.String(sourceKey),
		Marker:    latestKey,
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int64(maxKeys),
	})
	if err != nil {
		return nil, nil, formatError(err)
	}

	for i := range objs.CommonPrefixes {
		files = append(files, *objs.CommonPrefixes[i].Prefix)
	}
	for i := range objs.Contents {
		files = append(files, *objs.Contents[i].Key)
	}

	return files, objs.NextMarker, nil
}

func (x Bucket) Presign(keyOrUrl string, expire time.Duration) (putUrl string, getUrl string, err error) {
	bucketName, appId, sourceKey, err := x.parse(keyOrUrl)
	if err != nil {
		bucketName = x.bucketName
		appId = x.appId
		sourceKey = keyOrUrl
	}

	request, _ := x.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(fmt.Sprintf("%s-%s", bucketName, appId)),
		Key:    aws.String(sourceKey),
	})
	putUrl, err = request.Presign(expire)
	if err != nil {
		return "", "", formatError(err)
	}

	request, _ = x.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("%s-%s", bucketName, appId)),
		Key:    aws.String(sourceKey),
	})
	getUrl, err = request.Presign(expire)
	if err != nil {
		return "", "", formatError(err)
	}
	return putUrl, getUrl, err
}

var configBytes = []byte(`
host: ${COS_HOST}
bucket_name: ${COS_BUCKET_NAME}
app_id: ${COS_APP_ID}
region: ${COS_REGION}
secret_id: ${COS_SECRET_ID}
secret_key: ${COS_SECRET_KEY}
token: ""
`)

func NewBucketFromEnv(endpoint Endpoint, opts ...Option) (*Bucket, error) {
	cfg := &Configuration{}
	if err := config.Unmarshal(configBytes, cfg); err != nil {
		log.WithError(err).Fatal()
	}
	return NewBucket(cfg, endpoint, opts...)
}
