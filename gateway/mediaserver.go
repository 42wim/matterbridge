package gateway

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/sirupsen/logrus"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type mediaServer interface {
	handleFilesUpload(fi *config.FileInfo) (string, error)
}

type commonMediaServer struct {
	logger *logrus.Entry
}

type httpPutMediaServer struct {
	commonMediaServer

	httpUploadPath     string
	httpDownloadPrefix string
}

type localMediaServer struct {
	commonMediaServer

	localPath          string
	httpDownloadPrefix string
}

type minioMediaServer struct {
	commonMediaServer

	minio          *minio.Client
	bucket         string
	uploadPrefix   string
	downloadPrefix string

	ctx context.Context
}

var _ mediaServer = (*httpPutMediaServer)(nil)
var _ mediaServer = (*localMediaServer)(nil)
var _ mediaServer = (*minioMediaServer)(nil)

func createMediaServer(bg *config.BridgeValues, logger *logrus.Entry) (mediaServer, error) {
	if bg.General.MediaServerUpload == "" && bg.General.MediaDownloadPath == "" {
		return nil, nil //  we don't have a attachfield or we don't have a mediaserver configured return
	}

	if bg.General.MediaServerUpload != "" {
		parsed, err := url.Parse(bg.General.MediaServerUpload)
		if err != nil {
			return nil, fmt.Errorf("failed parsing mediaServerUpload URL: %w", err)
		}

		if parsed.Scheme == "http" || parsed.Scheme == "https" {
			return &httpPutMediaServer{
				commonMediaServer: commonMediaServer{
					logger: logger,
				},

				httpUploadPath:     bg.General.MediaServerUpload,
				httpDownloadPrefix: bg.General.MediaServerDownload,
			}, nil
		}

		if parsed.Scheme == "minio" {
			optionsFromURL := parsed.Query()
			secretAccessKey, _ := parsed.User.Password()
			pathSplitted := strings.Split(strings.TrimLeft(parsed.Path, "/"), "/")

			useSSL, err := strconv.ParseBool(optionsFromURL.Get("useSSL"))
			if err != nil {
				logger.Warn("error while parsing useSSL boolean, assuming false: ", err)
				useSSL = false
			}

			if len(pathSplitted) == 0 {
				return nil, fmt.Errorf("no bucket specified")
			}

			bucketName := pathSplitted[0]
			uploadPrefix := strings.Join(pathSplitted[1:], "/")

			ctx := context.Background()

			minioClient, err := minio.New(parsed.Host, &minio.Options{
				Creds:  credentials.NewStaticV4(parsed.User.Username(), secretAccessKey, ""),
				Secure: useSSL,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to initialize minio client: %w", err)
			}

			logger.WithFields(logrus.Fields{
				"bucket":       bucketName,
				"uploadPrefix": uploadPrefix,
			}).Debug("configured minio client")

			exist, err := minioClient.BucketExists(ctx, bucketName)
			if err != nil {
				return nil, fmt.Errorf("failed checking if bucket exists: %w", err)
			}
			if !exist {
				return nil, fmt.Errorf("specified bucket does not exists")
			}

			return &minioMediaServer{
				commonMediaServer: commonMediaServer{
					logger: logger,
				},

				ctx:   ctx,
				minio: minioClient,

				bucket:         bucketName,
				uploadPrefix:   uploadPrefix,
				downloadPrefix: bg.General.MediaServerDownload,
			}, nil
		}

		return nil, fmt.Errorf("unknown schema (protocol) for mediaServerUpload: '%s'", parsed.Scheme)
	}

	if bg.General.MediaDownloadPath != "" {
		return &localMediaServer{
			commonMediaServer: commonMediaServer{
				logger: logger,
			},

			localPath:          bg.General.MediaDownloadPath,
			httpDownloadPrefix: bg.General.MediaServerDownload,
		}, nil
	}

	return nil, nil // never reached
}

// handleFilesUpload which uses MediaServerUpload configuration to upload the file via HTTP PUT request.
// Returns error on failure.
func (h *httpPutMediaServer) handleFilesUpload(fi *config.FileInfo) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	// Use MediaServerUpload. Upload using a PUT HTTP request and basicauth.
	sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))[:8] //nolint:gosec
	url := h.httpUploadPath + "/" + sha1sum + "/" + fi.Name

	req, err := http.NewRequest("PUT", url, bytes.NewReader(*fi.Data))
	if err != nil {
		return "", fmt.Errorf("mediaserver upload failed, could not create request: %#v", err)
	}

	h.logger.Debugf("mediaserver upload url: %s", url)

	req.Header.Set("Content-Type", "binary/octet-stream")
	_, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("mediaserver upload failed, could not Do request: %#v", err)
	}

	return h.httpDownloadPrefix + "/" + sha1sum + "/" + fi.Name, nil
}

// handleFilesUpload which uses MediaServerPath configuration, places the file on the current filesystem.
// Returns error on failure.
func (h *localMediaServer) handleFilesUpload(fi *config.FileInfo) (string, error) {
	sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))[:8] //nolint:gosec
	dir := h.localPath + "/" + sha1sum
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("mediaserver path failed, could not mkdir: %s %#v", err, err)
	}

	path := dir + "/" + fi.Name
	h.logger.Debugf("mediaserver path placing file: %s", path)

	err = ioutil.WriteFile(path, *fi.Data, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("mediaserver path failed, could not writefile: %s %#v", err, err)
	}

	return h.httpDownloadPrefix + "/" + sha1sum + "/" + fi.Name, nil
}

// handleFilesUpload which uploads media to minio compatible server (S3)
// Returns error on failure.
func (h *minioMediaServer) handleFilesUpload(fi *config.FileInfo) (string, error) {
	sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))[:8]
	url := h.uploadPrefix + "/" + sha1sum + "/" + fi.Name
	objectSize := int64(len(*fi.Data)) // TODO: Using this, since we got this in memory anyway. Would be nicer to use fi.Size, but it is 0

	info, err := h.minio.PutObject(h.ctx, h.bucket, url, bytes.NewReader(*fi.Data), objectSize, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return "", fmt.Errorf("mediaserver putfile failed: %w", err)
	}

	h.logger.Debugf("successfully uploaded %v, etag: %v", url, info.ETag)
	return h.downloadPrefix + url, nil
}
