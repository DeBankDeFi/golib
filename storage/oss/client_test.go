// +build integration

package oss_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DeBankDeFi/golib/storage/oss"

	"github.com/stretchr/testify/require"
)

var client *oss.AuthClient

func init() {
	client = oss.NewClient(&oss.ClientOpt{
		Bucket:          os.Getenv("ALI_OSS_BUCKET"),
		AccessKeyID:     os.Getenv("ALI_OSS_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("ALI_OSS_SECRET_KEY"),
		ClientTimeout:   50 * time.Second,
		MaxRetry:        5,
	})
}

func TestClient_CRUD(t *testing.T) {
	// put
	err := client.PutObjectWithTTL(context.Background(), "test/test-key", []byte("hello-world"), time.Minute)
	require.NoError(t, err)

	// get
	data, err := client.GetObject(context.Background(), "test/test-key")
	require.NoError(t, err)
	t.Log(string(data))

	// list
	objs, err := client.ListObjects(context.Background(), "test/test-key", 1000)
	require.NoError(t, err)
	for _, obj := range objs {
		fmt.Printf("Name: %s \t Latest Modified: %s \t Size: %d\n", obj.Name, obj.LatestModified.String(), obj.Size)
	}
}
