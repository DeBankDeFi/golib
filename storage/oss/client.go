package oss

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"sort"
	"time"

	"github.com/DeBankDeFi/golib/syserror"
	"github.com/DeBankDeFi/golib/util"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const (
	DefaultOSSACL       = oss.ACLPublicRead
	DefaultStorageClass = oss.StorageStandard
)

type Client struct {
	oss *oss.Client
	opt *ClientOpt
}

// ClientOpts
type ClientOpt struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string

	ClientTimeout  time.Duration
	MaxRetry       int
	ForcePathStyle bool
}

func (c *Client) getBucket(ctx context.Context) (bucket *oss.Bucket, err error) {
	bucket, err = c.oss.Bucket(c.opt.Bucket)

	if err != nil {
		return nil, syserror.New(util.GetTraceIDFromContext(ctx), "OSS_OBJECT_GET_BUCKET", err.Error(), map[string]interface{}{
			"Bucket": c.opt.Bucket,
		})
	}

	return
}

func (c *Client) putObject(ctx context.Context, key string, content []byte) (*oss.Bucket, error) {
	body := bytes.NewReader(content)

	bucket, err := c.getBucket(ctx)
	if err != nil {
		return nil, err
	}

	storageType := oss.ObjectStorageClass(DefaultStorageClass)
	acl := oss.ObjectACL(DefaultOSSACL)

	return bucket, bucket.PutObject(key, body, storageType, acl)
}

func (c *Client) PutObject(ctx context.Context, key string, content []byte) (err error) {
	bucket, err := c.putObject(ctx, key, content)

	if bucket != nil && err != nil {
		return syserror.New(util.GetTraceIDFromContext(ctx), "OSS_OBJECT_PUT", err.Error(), map[string]interface{}{
			"Bucket": bucket.BucketName,
			"Key":    key,
		})
	} else if bucket == nil { // bucket error
		return err
	}

	return nil
}

func (c *Client) PutObjectWithTTL(ctx context.Context, key string, content []byte, ttl time.Duration) error {
	// TODO: implement me
	return c.PutObject(ctx, key, content)
}

func (c *Client) HeadObject(ctx context.Context, key string) (bool, error) {
	bucket, err := c.getBucket(ctx)
	if err != nil {
		return false, err
	}

	isExist, err := bucket.IsObjectExist(key)
	if err != nil || !isExist {
		return false, syserror.New(util.GetTraceIDFromContext(ctx), "OSS_OBJECT_HEAD", err.Error(), map[string]interface{}{
			"Bucket": c.opt.Bucket,
			"Key":    key,
		})
	}

	return true, nil
}

func (c *Client) getObject(ctx context.Context, key string) (rc io.ReadCloser, err error) {
	bucket, err := c.getBucket(ctx)
	if err != nil {
		return nil, err
	}

	output, err := bucket.GetObject(key)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// GetObject gets object content from oss.
func (c *Client) GetObject(ctx context.Context, key string) (body []byte, err error) {
	rc, err := c.getObject(ctx, key)
	if err != nil {
		return nil, syserror.New(util.GetTraceIDFromContext(ctx), "OSS_OBJECT_GET", err.Error(), map[string]interface{}{
			"Bucket": c.opt.Bucket,
			"Key":    key,
		})
	}
	defer rc.Close()

	body, err = ioutil.ReadAll(rc)
	if err != nil {
		return nil, syserror.New(util.GetTraceIDFromContext(ctx), "OSS_OBJECT_CONTENT_READ", err.Error(), map[string]interface{}{
			"Bucket": c.opt.Bucket,
			"Key":    key,
		})
	}

	return
}

func (c *Client) GetObjectStream(ctx context.Context, key string) (body io.ReadCloser, err error) {
	return c.getObject(ctx, key)
}

func (c *Client) listObjects(ctx context.Context, keyPrefix string, limit int) (Objects, error) {
	bucket, err := c.getBucket(ctx)
	if err != nil {
		return nil, err
	}

	var data oss.ListObjectsResult
	var objects Objects
	if limit > 0 {
		// data.MaxKeys = aws.Int64(int64(limit))
		data, err = bucket.ListObjects(oss.Prefix(keyPrefix), oss.MaxKeys(limit))
	} else {
		data, err = bucket.ListObjects(oss.Prefix(keyPrefix))
	}
	if err != nil {
		return nil, err
	}

	objects = data.Objects
	sort.Sort(objects)

	return objects, nil
}

// ListObjects lists all objects with key prefix.
func (c *Client) ListObjects(ctx context.Context, keyPrefix string, limit int) ([]Object, error) {
	objs, err := c.listObjects(ctx, keyPrefix, limit)
	if err != nil {
		return nil, syserror.New(util.GetTraceIDFromContext(ctx), "OSS_OBJECT_LIST", err.Error(), map[string]interface{}{
			"Bucket":    c.opt.Bucket,
			"KeyPrefix": keyPrefix,
		})
	}
	ossObjs := make([]Object, 0, len(objs))

	for _, obj := range objs {
		ossObjs = append(ossObjs, Object{
			Name:           obj.Key,
			LatestModified: obj.LastModified,
			Size:           obj.Size,
		})
	}

	return ossObjs, nil
}

type Object struct {
	Name           string    `json:"name"`
	LatestModified time.Time `json:"latest_modified"`
	Size           int64     `json:"size"`
}
type Objects []oss.ObjectProperties

func (oo Objects) Len() int { return len(oo) }
func (oo Objects) Less(i, j int) bool {
	im := oo[i].LastModified
	jm := oo[j].LastModified
	return im.After(jm)
}
func (oo Objects) Swap(i, j int) { oo[i], oo[j] = oo[j], oo[i] }

// UnAuthClient a s3 client that need authentication.
type AuthClient struct {
	*Client
}

// NewAuthClient creates a new auth client.
func NewClient(opt *ClientOpt) *AuthClient {
	ossClient, _ := oss.New(
		opt.Endpoint,
		opt.AccessKeyID,
		opt.SecretAccessKey,
		oss.Timeout(int64(opt.ClientTimeout), int64(opt.ClientTimeout)),
	)

	return &AuthClient{
		Client: &Client{
			oss: ossClient,
			opt: opt,
		},
	}
}
