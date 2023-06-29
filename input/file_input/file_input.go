package file_input

import (
	"log"
	"sync"
	"time"

	"github.com/childe/gohangout/topology"
)

type FileInput struct {
	paths            []string
	localStorageFile string
	localStorage     *LocalStorage
	watcher          *Watcher
	channel          chan string
}

type fileOffsetsItem struct {
	Offsets int
	Lines   int
}

func (fi *FileInput) getFileOffsetsItem(filePath string) *fileOffsetsItem {
	// 使用文件路径Hash作为key
	key := MD5(filePath, 16)
	value := []interface{}{float64(0), float64(0)}
	value = fi.localStorage.Get(key, value).([]interface{})

	item := &fileOffsetsItem{}
	item.Offsets = int(value[0].(float64))
	item.Lines = int(value[1].(float64))

	return item
}

func (fi *FileInput) setFileOffsetsItem(filePath string, fileOffsetsItem *fileOffsetsItem) {
	// 使用文件路径Hash作为key
	key := MD5(filePath, 16)
	value := []interface{}{float64(0), float64(0)}
	value[0] = fileOffsetsItem.Offsets
	value[1] = fileOffsetsItem.Lines

	fi.localStorage.Set(key, value)
}

func (fi *FileInput) delFileOffsetsItem(filePath string) {
	// 使用文件路径Hash作为key
	key := MD5(filePath, 16)
	fi.localStorage.Del(key)
}

// StartReaders 启动Reader
func (fi *FileInput) StartReaders() {
	readers := make(map[string]*Reader)
	readersLock := sync.Mutex{}

	go func() {
		for {
			change := fi.watcher.GetChange()

			if change == nil {
				time.Sleep(time.Second * 2)
				continue
			}

			readersLock.Lock()

			switch change.FileChangeType {
			case FileChangeAdd:
				readers[change.FilePath] = fi.StartNewReader(change.FilePath)
			case FileChangeRemove:
				readers[change.FilePath].Stop()
				delete(readers, change.FilePath)

				// 删除文件行数记录
				fi.delFileOffsetsItem(change.FilePath)
			}

			readersLock.Unlock()
		}
	}()
}

// StartNewReader 启动新Reader
func (fi *FileInput) StartNewReader(filePath string) *Reader {
	fileOffsetsItem := fi.getFileOffsetsItem(filePath)

	reader := NewReader(filePath, fileOffsetsItem.Offsets, fileOffsetsItem.Lines)

	// 启动Watch线程
	go func() {
		Logf("Watching file \"%s\"\n", filePath)

		err := reader.Watch()

		if err != nil {
			Log(err)
		}

		Logf("Abort watch file \"%s\"\n", filePath)
	}()

	// 启动读取线程
	go func() {
		for {
			line, isOpen := <-reader.Channel

			// 通道被关闭
			if !isOpen {
				break
			}

			fi.channel <- line

			// 记录最新行数
			fileOffsetsItem.Lines = reader.Lines()
			fileOffsetsItem.Offsets = reader.Offsets()
			fi.setFileOffsetsItem(filePath, fileOffsetsItem)
		}
	}()

	return reader
}

func NewFileInput(config map[interface{}]interface{}) topology.Input {
	input := &FileInput{
		paths:            make([]string, 0),
		localStorageFile: config["local_storage_file"].(string),
		channel:          make(chan string),
	}

	for _, v := range config["paths"].([]interface{}) {
		input.paths = append(input.paths, v.(string))
	}

	input.Init()

	return input
}

// Init 插件初始化
func (fi *FileInput) Init() {
	localStorage, err := NewLocalStorage(fi.localStorageFile)

	if err != nil {
		log.Fatalln(err)
	}

	fi.localStorage = localStorage
	fi.watcher = NewWatcher()

	// 添加Path
	fi.watcher.AddPath(fi.paths...)
	// 开始监控文件
	fi.watcher.WatchFiles()

	// 根据扫描文件启动Reader
	fi.StartReaders()
}

func (fi *FileInput) ReadOneEvent() map[string]interface{} {
	return map[string]interface{}{"message": <-fi.channel}
}

func (fi *FileInput) Shutdown() {
	defer func() {
		if err := fi.localStorage.Save(); err != nil {
			log.Fatalln(err)
		}
	}()
}
