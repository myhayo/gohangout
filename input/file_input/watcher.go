package file_input

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileChange struct {
	FileChangeType FileChangeType
	FilePath       string
}

type FileChangeType int8

const FileChangeAdd FileChangeType = 1
const FileChangeRemove FileChangeType = 2

type Watcher struct {
	paths []string
	queue *Queue

	pathLock sync.Mutex
}

// AddPath 添加Path
func (w *Watcher) AddPath(path ...string) {
	w.pathLock.Lock()
	w.paths = append(w.paths, path...)
	w.pathLock.Unlock()
}

// GetChange 获取一个改动
func (w *Watcher) GetChange() *FileChange {
	v := w.queue.Get()

	if v == nil {
		return nil
	}

	return v.(*FileChange)
}

// ScanFiles 扫描文件
func (w *Watcher) ScanFiles() map[string]bool {
	// 读取文件列表
	filePathMap := make(map[string]bool)

	w.pathLock.Lock()
	paths := w.paths
	w.pathLock.Unlock()

	for _, item := range paths {
		abs, err := filepath.Abs(item)

		if err != nil {
			Log(err)
			continue
		}

		paths, err := filepath.Glob(abs)

		if err != nil {
			Log(err)
			continue
		}

		for _, p := range paths {
			// 检查是否文件夹
			fs, err := os.Stat(p)

			if err != nil {
				Log(err)
				continue
			}

			// 仅处理非文件夹
			if fs.IsDir() {
				Logf("Don't Support Dir \"%s\"\n", p)
				continue
			}

			// 加入文件map
			filePathMap[p] = true
		}
	}

	return filePathMap
}

// WatchFiles 监控文件
func (w *Watcher) WatchFiles() {
	go func() {
		lastFiles := make(map[string]bool)

		for {
			files := w.ScanFiles()

			// 对比文件
			left := make([]string, 0)
			right := make([]string, 0)
			for file := range files {
				if _, exists := lastFiles[file]; !exists {
					left = append(left, file)
				}
			}
			for file := range lastFiles {
				if _, exists := files[file]; !exists {
					right = append(right, file)
				}
			}

			// ADD
			for _, file := range left {
				Logf("Add watch \"%s\"\n", file)
				lastFiles[file] = true
				w.queue.Add(&FileChange{
					FilePath:       file,
					FileChangeType: FileChangeAdd,
				})
			}
			// REMOVE
			for _, file := range right {
				Logf("Remove watch \"%s\"\n", file)
				delete(lastFiles, file)
				w.queue.Add(&FileChange{
					FilePath:       file,
					FileChangeType: FileChangeRemove,
				})
			}

			<-time.After(time.Second * 2)
		}
	}()
}

func NewWatcher() *Watcher {
	w := Watcher{
		queue: &Queue{},
	}

	return &w
}
