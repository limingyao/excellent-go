package cos

import (
	"context"
	"strings"
	"testing"
	"time"
)

func testBucket(t *testing.T, bucket *Bucket) {
	url, err := bucket.Put(context.Background(), "hello", []byte("hello"))
	t.Log(url, err)

	buffer, err := bucket.Get(context.Background(), url)
	t.Log(string(buffer), err)

	url, err = bucket.Copy(context.Background(), url, "hello-tmp")
	t.Log(url, err)

	files, key, err := bucket.List(context.Background(), url[0:strings.LastIndex(url, "/")+1], nil, 1000)
	t.Log(files, key, err)

	buffer, err = bucket.Download(context.Background(), url)
	t.Log(string(buffer), err)

	err = bucket.Delete(context.Background(), url)
	t.Log(err)

	files, key, err = bucket.List(context.Background(), url[0:strings.LastIndex(url, "/")+1], nil, 1000)
	t.Log(files, key, err)

	putUrl, getUrl, err := bucket.Presign("hello", time.Minute)
	t.Log(putUrl)
	t.Log(getUrl)
	t.Log(err)
}

func TestPublicBucket(t *testing.T) {
	bucket, err := NewBucket(
		"",
		"test",
		"12345678",
		"ap-beijing",
		"id",
		"key",
		"",
		QCloudEndpoint,
	)
	if err != nil {
		t.Error(err)
		return
	}

	testBucket(t, bucket)
}

func TestPrivateBucket(t *testing.T) {
	bucket, err := NewBucket(
		"http://dev.machine:10000",
		"test",
		"1",
		"ap-local",
		"id",
		"key",
		"",
		RawEndpoint,
		WithPathStyle(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	testBucket(t, bucket)
}

func TestParseVirtualHostedStyle(t *testing.T) {
	urls := []string{
		// fail
		"https://test-bj-01-12345678.cos.ap-beijing.myqcloud.com",
		"http://test-12345678",
		// ok
		"http://bucket-capture-face-7day-1/test/20191010.png",
		"https://test-bj-01-12345678.cos.ap-beijing.myqcloud.com/test.cos",
		"test-bj-01-12345678.cos.ap-beijing.myqcloud.com/test.cos",
		"https://test-12345678.cos.ap-beijing.myqcloud.com/test#20191010",
		"https://test-12345678.cos.ap-beijing.myqcloud.com/test.cos?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE",
	}
	for i := range urls {
		t.Log(parseVirtualHostedStyle(urls[i]))
	}
}

func TestParsePathStyle(t *testing.T) {
	urls := []string{
		"http://127.0.0.1:9000/test-1/path/key1?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE",
		"127.0.0.1:9000/test-1/path/key1?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE",
		"http://test-1/path/key1?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE",
		"test-1/path/key1?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE",
	}
	for i := range urls {
		t.Log(parsePathStyle(urls[i]))
	}
}
