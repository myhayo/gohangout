package file_input

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type LocalStorage struct {
	path     string
	saved    atomic.Value
	data     sync.Map
	saveLock sync.Mutex
}

// 读取
func (l *LocalStorage) load() error {
	var err error

	// 初始化
	l.saved.Store(true)

	// 绝对路径格式化
	l.path, err = filepath.Abs(l.path)

	if err != nil {
		return err
	}

	// 打开文件
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDONLY, fs.ModePerm)

	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	// 读取内容
	fileData, err := io.ReadAll(f)

	if len(fileData) == 0 {
		return err
	}

	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(fileData, &data)

	// 写入sync.Map
	for k, v := range data {
		l.data.Store(k, v)
	}

	return err
}

// Save 保存
func (l *LocalStorage) Save() error {
	l.saveLock.Lock()

	data := make(map[string]interface{})

	// 读取sync.Map
	l.data.Range(func(k, v interface{}) bool {
		data[k.(string)] = v

		return true
	})

	fileData, err := json.Marshal(&data)

	if err != nil {
		l.saveLock.Unlock()

		return err
	}

	// 写入临时文件
	tmpFilepath := filepath.Dir(l.path) + "/." + filepath.Base(l.path) + ".tmp"
	err = os.WriteFile(tmpFilepath, fileData, fs.ModePerm)

	if err != nil {
		_ = os.Remove(tmpFilepath)
		l.saveLock.Unlock()

		return err
	}

	err = os.Rename(tmpFilepath, l.path)

	if err != nil {
		l.saveLock.Unlock()

		return err
	}

	l.saved.Store(true)
	l.saveLock.Unlock()

	return nil
}

// Get 获取值
func (l *LocalStorage) Get(key string, defaultValue interface{}) interface{} {
	v, _ := l.data.LoadOrStore(key, defaultValue)

	return v
}

// Set 写入值
func (l *LocalStorage) Set(key string, value interface{}) {
	l.data.Store(key, value)
	l.saved.Store(false)
}

// Del 删除值
func (l *LocalStorage) Del(key string) {
	l.data.Delete(key)
}

// NewLocalStorage 创建LocalStorage
func NewLocalStorage(path string) (*LocalStorage, error) {
	l := LocalStorage{
		path: path,
	}

	err := l.load()

	if err != nil {
		return nil, err
	}

	// 自动保存线程
	go func() {
		for {
			// 若未保存
			if !l.saved.Load().(bool) {
				err := l.Save()

				if err != nil {
					Log(err)
				}
			}

			time.Sleep(time.Second)
		}
	}()

	return &l, nil
}
