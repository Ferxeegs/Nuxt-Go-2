// Package uploader validates and persists uploaded files to local disk or MinIO/S3.
package uploader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	appconfig "go-cp-boilerplate/internal/config"
)

var (
	ErrExtensionNotAllowed = errors.New("file extension not allowed")
	ErrMimeNotAllowed      = errors.New("file content type not allowed")
	ErrFileTooLarge        = errors.New("file exceeds maximum allowed size")
	ErrS3NotConfigured     = errors.New("s3 storage is not configured")
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".pdf":  true,
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

// Saved describes a persisted file.
type Saved struct {
	FileName     string
	StoredName   string
	RelativePath string // object key / relative path used in public URL
	MimeType     string
	Size         int64
	PublicURL    string // absolute or site-relative URL for clients
}

// Uploader stores files on local disk and optionally on MinIO/S3.
type Uploader struct {
	BaseDir   string
	MaxSizeMB int64

	s3Client       *s3.Client
	s3Bucket       string
	s3PublicBucket string
	// publicBase is the URL prefix returned to clients for S3 objects
	// (e.g. "/storage" proxied by nginx, or full http://localhost:9020/bucket).
	publicBase string
	localBase  string // e.g. "/uploads"
}

func New(baseDir string, maxSizeMB int64) *Uploader {
	return &Uploader{
		BaseDir:    baseDir,
		MaxSizeMB:  maxSizeMB,
		localBase:  "/uploads",
		publicBase: "/uploads",
	}
}

// NewFromConfig builds an uploader that prefers MinIO when S3_ENDPOINT_URL is set.
func NewFromConfig(cfg *appconfig.Config) (*Uploader, error) {
	u := New(cfg.UploadDir, cfg.MaxUploadSizeMB)
	if strings.TrimSpace(cfg.S3EndpointURL) == "" {
		return u, nil
	}

	bucket := strings.TrimSpace(cfg.S3PublicBucket)
	if bucket == "" {
		bucket = strings.TrimSpace(cfg.S3Bucket)
	}
	if bucket == "" {
		return nil, fmt.Errorf("S3_PUBLIC_BUCKET or S3_BUCKET is required when S3_ENDPOINT_URL is set")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(strings.TrimRight(cfg.S3EndpointURL, "/"))
		o.UsePathStyle = cfg.S3ForcePathStyle
	})

	u.s3Client = client
	u.s3Bucket = strings.TrimSpace(cfg.S3Bucket)
	u.s3PublicBucket = bucket
	// Prefer same-origin nginx proxy path; fallback to public MinIO endpoint.
	if pub := strings.TrimRight(strings.TrimSpace(cfg.S3PublicEndpointURL), "/"); pub != "" {
		// If public endpoint already includes the bucket path, use as-is;
		// otherwise append bucket for path-style URLs.
		if strings.HasSuffix(pub, "/"+bucket) || strings.Contains(pub, "/"+bucket+"/") {
			u.publicBase = pub
		} else if strings.HasPrefix(pub, "http://") || strings.HasPrefix(pub, "https://") {
			u.publicBase = pub + "/" + bucket
		} else {
			u.publicBase = pub
		}
	} else {
		u.publicBase = "/storage"
	}
	return u, nil
}

func (u *Uploader) UsesS3() bool {
	return u != nil && u.s3Client != nil
}

func (u *Uploader) PublicURLPrefix() string {
	if u.UsesS3() {
		return u.publicBase
	}
	return u.localBase
}

// Save validates and stores the file. When S3 is configured, objects go to the
// public bucket so settings images (and other web assets) are anonymously readable.
func (u *Uploader) Save(fh *multipart.FileHeader, module string) (*Saved, error) {
	maxBytes := u.MaxSizeMB * 1024 * 1024
	if fh.Size > maxBytes {
		return nil, ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedExtensions[ext] {
		return nil, ErrExtensionNotAllowed
	}

	src, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()

	head := make([]byte, 512)
	n, _ := src.Read(head)
	mimeType := http.DetectContentType(head[:n])
	if !allowedMimeTypes[mimeType] {
		return nil, ErrMimeNotAllowed
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	now := time.Now()
	module = sanitizeModule(module)
	relDir := filepath.ToSlash(filepath.Join(module, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month())))
	storedName := fmt.Sprintf("%s-%d%s", uuid.NewString(), now.Unix(), ext)
	relPath := relDir + "/" + storedName

	if u.UsesS3() {
		body, err := io.ReadAll(src)
		if err != nil {
			return nil, err
		}
		_, err = u.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket:      aws.String(u.s3PublicBucket),
			Key:         aws.String(relPath),
			Body:        bytes.NewReader(body),
			ContentType: aws.String(mimeType),
		})
		if err != nil {
			return nil, fmt.Errorf("s3 put object: %w", err)
		}
		return &Saved{
			FileName:     fh.Filename,
			StoredName:   storedName,
			RelativePath: relPath,
			MimeType:     mimeType,
			Size:         fh.Size,
			PublicURL:    strings.TrimRight(u.publicBase, "/") + "/" + relPath,
		}, nil
	}

	fullDir := filepath.Join(u.BaseDir, filepath.FromSlash(relDir))
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		return nil, err
	}
	fullPath := filepath.Join(fullDir, storedName)
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = dst.Close() }()
	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	return &Saved{
		FileName:     fh.Filename,
		StoredName:   storedName,
		RelativePath: relPath,
		MimeType:     mimeType,
		Size:         fh.Size,
		PublicURL:    strings.TrimRight(u.localBase, "/") + "/" + relPath,
	}, nil
}

// DeleteObject removes an S3 object when S3 is configured. Local disk cleanup
// is best-effort and ignored if the file is missing.
func (u *Uploader) DeleteObject(relativePath string) error {
	relativePath = strings.TrimPrefix(filepath.ToSlash(relativePath), "/")
	if relativePath == "" {
		return nil
	}
	if u.UsesS3() {
		_, err := u.s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(u.s3PublicBucket),
			Key:    aws.String(relativePath),
		})
		return err
	}
	full := filepath.Join(u.BaseDir, filepath.FromSlash(relativePath))
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func sanitizeModule(module string) string {
	module = strings.TrimSpace(module)
	if module == "" {
		return "general"
	}
	module = strings.ReplaceAll(module, "..", "")
	module = strings.ReplaceAll(module, "\\", "/")
	module = strings.Trim(module, "/")
	if module == "" {
		return "general"
	}
	return module
}
