package s3

import (
	"context"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/storage-bridge/core/pkg/storage"
)

type Provider struct {
	client *s3.Client
	bucket string
}

func New(client *s3.Client, bucket string) *Provider {
	return &Provider{
		client: client,
		bucket: bucket,
	}
}

func (p *Provider) Name() string {
	return "s3"
}

func (p *Provider) Stat(ctx context.Context, path string) (*storage.Entry, error) {
	path = strings.TrimPrefix(path, "/")
	
	if path == "" {
		return &storage.Entry{Path: "", IsDir: true, ModTime: time.Now()}, nil
	}

	out, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})
	
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) || strings.Contains(err.Error(), "NotFound") {
			// S3 doesn't have real directories, check if there's a prefix
			listOut, err := p.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket:  aws.String(p.bucket),
				Prefix:  aws.String(path + "/"),
				MaxKeys: aws.Int32(1),
			})
			if err == nil && len(listOut.Contents) > 0 {
				return &storage.Entry{Path: path, IsDir: true, ModTime: time.Now()}, nil
			}
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return &storage.Entry{
		Path:    path,
		IsDir:   false,
		Size:    aws.ToInt64(out.ContentLength),
		ModTime: aws.ToTime(out.LastModified),
	}, nil
}

type s3Iterator struct {
	p        *Provider
	path     string
	paginator *s3.ListObjectsV2Paginator
	page     *s3.ListObjectsV2Output
	idx      int
}

func (it *s3Iterator) Next(ctx context.Context) (*storage.Entry, error) {
	if it.page == nil || it.idx >= len(it.page.Contents)+len(it.page.CommonPrefixes) {
		if !it.paginator.HasMorePages() {
			return nil, io.EOF
		}
		
		page, err := it.paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		
		it.page = page
		it.idx = 0
		
		if len(it.page.Contents) == 0 && len(it.page.CommonPrefixes) == 0 {
			return nil, io.EOF
		}
	}
	
	if it.idx < len(it.page.CommonPrefixes) {
		prefix := aws.ToString(it.page.CommonPrefixes[it.idx].Prefix)
		it.idx++
		return &storage.Entry{
			Path:  prefix[:len(prefix)-1], // strip trailing slash
			IsDir: true,
		}, nil
	}
	
	objIdx := it.idx - len(it.page.CommonPrefixes)
	obj := it.page.Contents[objIdx]
	it.idx++
	
	return &storage.Entry{
		Path:    aws.ToString(obj.Key),
		IsDir:   false,
		Size:    aws.ToInt64(obj.Size),
		ModTime: aws.ToTime(obj.LastModified),
	}, nil
}

func (p *Provider) List(ctx context.Context, path string) (storage.Iterator, error) {
	path = strings.TrimPrefix(path, "/")
	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	paginator := s3.NewListObjectsV2Paginator(p.client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(p.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})

	return &s3Iterator{
		p:         p,
		path:      prefix,
		paginator: paginator,
	}, nil
}

func (p *Provider) Get(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error) {
	path = strings.TrimPrefix(path, "/")
	
	input := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	}
	
	if offset > 0 || length >= 0 {
		end := ""
		if length >= 0 {
			end = fmt.Sprintf("%d", offset+length-1)
		}
		input.Range = aws.String(fmt.Sprintf("bytes=%d-%s", offset, end))
	}
	
	out, err := p.client.GetObject(ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) || strings.Contains(err.Error(), "NoSuchKey") {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	
	return out.Body, nil
}

func (p *Provider) Put(ctx context.Context, path string, in io.Reader, size int64, modTime time.Time) error {
	path = strings.TrimPrefix(path, "/")
	
	uploader := manager.NewUploader(p.client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // 5MB parts
	})
	
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
		Body:   in,
	})
	
	return err
}

func (p *Provider) Remove(ctx context.Context, path string) error {
	path = strings.TrimPrefix(path, "/")
	
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})
	
	return err
}

func (p *Provider) SpaceUsed(ctx context.Context, path string) (int64, error) {
	path = strings.TrimPrefix(path, "/")
	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	paginator := s3.NewListObjectsV2Paginator(p.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucket),
		Prefix: aws.String(prefix),
	})

	var total int64
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, err
		}
		for _, obj := range page.Contents {
			total += aws.ToInt64(obj.Size)
		}
	}

	return total, nil
}

func (p *Provider) Mkdir(ctx context.Context, path string) error {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	
	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(p.prefix + path),
		Body:   bytes.NewReader([]byte{}),
	})
	return err
}

func (p *Provider) Move(ctx context.Context, src string, dest string) error {
	src = strings.TrimPrefix(src, "/")
	dest = strings.TrimPrefix(dest, "/")
	
	srcKey := p.prefix + src
	destKey := p.prefix + dest
	
	// Try to copy single object
	_, err := p.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(p.bucket),
		CopySource: aws.String(p.bucket + "/" + srcKey),
		Key:        aws.String(destKey),
	})
	
	if err == nil {
		// If single object copy succeeded, delete the original
		return p.Remove(ctx, src)
	}
	
	// If it failed, it might be a directory. We'd have to list and copy all objects.
	// For simplicity in this engine, we'll return an error that directory moves aren't fully supported in S3 without a recursive walk.
	return fmt.Errorf("directory move or copy error: %v", err)
}
