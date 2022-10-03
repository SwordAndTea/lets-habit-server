package service

import (
	"context"
	"net/url"
	"os"
	"path"
)

type ObjectStorage interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
	PutObject(ctx context.Context, key string, data []byte) error
	ObjectKeyToURL(key string) string
}

var currentObjectStorageImpl ObjectStorage

func GetObjectStorageExecutor() ObjectStorage {
	return currentObjectStorageImpl
}

/********************** Local Mock Object Storage Impl  ***********************/

type objectStorageImplLocalMock struct {
	urlPrefix        string
	localStorageRoot string
}

func InitObjectStorageWithLocalMockImpl(urlPrefix string, rootDir string) error {
	currentObjectStorageImpl = &objectStorageImplLocalMock{urlPrefix: urlPrefix, localStorageRoot: rootDir}
	return nil
}

func (s *objectStorageImplLocalMock) GetObject(ctx context.Context, key string) ([]byte, error) {
	fileData, err := os.ReadFile(path.Join(s.localStorageRoot, key))
	if err != nil {
		return nil, err
	}
	return fileData, nil
}

func (s *objectStorageImplLocalMock) PutObject(ctx context.Context, key string, data []byte) error {
	err := os.WriteFile(path.Join(s.localStorageRoot, key), data, 666)
	if err != nil {
		return err
	}
	return nil
}

func (s *objectStorageImplLocalMock) ObjectKeyToURL(key string) string {
	uri, _ := url.JoinPath(s.urlPrefix, s.localStorageRoot, key)
	return uri
}

/********************** Wechat Cloud Object Storage Impl  ***********************/

type objectStorageImplWechatCloud struct {
}