// Package storage gerencia a conexão com o S3
package storage

/*
* Aqui será implementado todo upload e download de arquivos usando awk-sdk
* */

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type S3Storage struct {
	client *s3.Client
	bucket string
	config S3Config
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	// Criar custom resolver para MinIO
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Configurar credenciais
	awsCfg := aws.Config{
		Region:                      cfg.Region,
		EndpointResolverWithOptions: customResolver,
		Credentials:                 credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // Necessário para MinIO
	})

	storage := &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		config: cfg,
	}

	// Verificar se bucket existe, se não, criar
	if err := storage.ensureBucket(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return storage, nil
}

func (s *S3Storage) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	// Gerar nome único para o arquivo
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102-150405"), uuid.New().String()[:8], ext)
	key := filepath.Join(path, filename)

	// Abrir arquivo
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Upload para S3
	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Retornar URL do arquivo
	return s.GetFileURL(key), nil
}

func (s *S3Storage) DeleteFile(fileURL string) error {
	// Extrair key da URL
	key := s.extractKeyFromURL(fileURL)

	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

func (s *S3Storage) GetFileURL(key string) string {
	// Para MinIO/S3, retornar URL direto
	scheme := "http"
	if s.config.UseSSL {
		scheme = "https"
	}

	endpoint := strings.TrimPrefix(s.config.Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	return fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint, s.bucket, key)
}

func (s *S3Storage) ensureBucket() error {
	// Verificar se bucket existe
	_, err := s.client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})

	if err == nil {
		// Bucket já existe
		return nil
	}

	// Criar bucket
	_, err = s.client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func (s *S3Storage) extractKeyFromURL(fileURL string) string {
	// Remove o prefixo da URL para obter apenas a key
	parts := strings.Split(fileURL, fmt.Sprintf("/%s/", s.bucket))
	if len(parts) > 1 {
		return parts[1]
	}
	return fileURL
}
