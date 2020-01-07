package buchiage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/wano/senju"
	"gopkg.in/fsnotify/fsnotify.v1"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	Name    string          `json:"name"`
	Matcher string          `json:"matcher"`
	Dest    string          `json:"dest"`
	Bucket  string          `json:"bucket"`
	Event   senju.EventName `json:"event_name"`
}

type ConfigList struct {
	ConfList []*Config `json:"conf_list"`
}

type buchiage struct {
	configList   []*Config
	s            *senju.Senju
	fileNameFunc func(string, *Config) (string, error)
	mu           *sync.RWMutex
	s3Client     s3iface.S3API
	logger Logger
}

func New(configList []*Config, s3Client s3iface.S3API) *buchiage {
	s := senju.New()

	b := &buchiage{
		configList: configList,
		s:          s,
		fileNameFunc: func(original string, config *Config) (string, error) {
			return original, nil
		},
		mu:       &sync.RWMutex{},
		s3Client: s3Client,
		logger: globalLogger,
	}

	return b
}

func (b *buchiage) SetLogger(l Logger) {
	b.logger = l
}

func (b *buchiage) SetFileNameFunc(f func(string, *Config) (string, error)) {
	b.mu.Lock()
	b.fileNameFunc = f
	b.mu.Unlock()
}

func (b *buchiage) fileName(original string, config *Config) (string, error) {
	defer b.mu.RUnlock()
	b.mu.RLock()
	return b.fileNameFunc(original, config)
}

func (b *buchiage) Run() error {
	for _, v := range b.configList {
		handler := &buchiageHandler{
			config: v,
			b:      b,
			logger: b.logger,
		}
		eventHandler := senju.NewEventHandler()
		eventHandler.SetHandler(handler.handler, v.Event)
		b.s.Add(v.Name, eventHandler)
	}
	b.logger.Info("buchiage start")
	return b.s.Run()
}

func (b *buchiage) Close() error {
	b.logger.Info("buchiage stopping..")
	return b.s.Close()
}

func (b *buchiage) uploadFile(target string, config *Config) (string, error) {
	fileInfo, err := os.Lstat(target)
	if err != nil {
		return "", err
	}

	uploadFileName, err := b.fileName(fileInfo.Name(), config)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s/%s", config.Dest, uploadFileName)

	f, err := os.Open(target)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := f.Close(); err != nil {
			b.logger.Error(err)
		}
	}()

	input := &s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(config.Bucket),
		Key:    aws.String(key),
	}
	if _, err := b.s3Client.PutObject(input); err != nil {
		return "", err
	}

	return key, nil

}

type buchiageHandler struct {
	config *Config
	b      *buchiage
	logger Logger
}

func (ba *buchiageHandler) handler(event fsnotify.Event) func(context.Context) error {
	return func(ctx context.Context) error {
		dir := filepath.Dir(event.Name)
		fileNames, err := filepath.Glob(fmt.Sprintf("%s/%s*", dir, ba.config.Matcher))
		if err != nil {
			return err
		}

		if len(fileNames) == 0 {
			return nil
		}

		var newFileInfo os.FileInfo
		for _, fileName := range fileNames {
			fileInfo, err := os.Stat(fileName)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return err
			}

			if newFileInfo == nil {
				newFileInfo = fileInfo
			} else if fileInfo.ModTime().After(newFileInfo.ModTime()) {
				newFileInfo = fileInfo
			}
		}

		if newFileInfo == nil {
			return nil
		}

		targetFile := fmt.Sprintf("%s/%s", dir, newFileInfo.Name())
		ba.logger.Info(fmt.Sprintf("upload_target_file: %s", targetFile))

		go func() {
			key, err := ba.b.uploadFile(targetFile, ba.config)
			if err != nil {
				ba.logger.Error(err)
			} else {
				ba.logger.Info(fmt.Sprintf("file uploaded. bucket: %s, key: %s", ba.config.Bucket, key))
			}
		}()
		return nil
	}
}
