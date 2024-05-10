package file_input

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"time"
)

type Reader struct {
	gzip    bool
	path    string
	stopped bool
	paused  bool
	pause   chan bool
	offsets int
	lines   int
	Channel chan string
}

// Stop 停止
func (r *Reader) Stop() {
	r.stopped = true
}

// Pause 暂停
func (r *Reader) Pause() {
	r.paused = true
}

// Continue 继续
func (r *Reader) Continue() {
	r.paused = false
	r.pause <- true
}

// Gzip 是否Gzip
func (r *Reader) Gzip() bool {
	return r.gzip
}

// Lines 当前行数
func (r *Reader) Lines() int {
	return r.lines
}

// Offsets 偏移量
func (r *Reader) Offsets() int {
	return r.offsets
}

// Watch 开始监听
func (r *Reader) Watch() error {
	f, err := os.Open(r.path)

	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	var ioReader io.Reader

	if r.gzip {
		gzipReader, err := gzip.NewReader(f)

		if err != nil {
			return err
		}

		defer func() { _ = gzipReader.Close() }()

		ioReader = gzipReader
	} else {
		ioReader = f
	}

	buf := bufio.NewReader(ioReader)
	_, _ = buf.Discard(r.offsets)
	var line []byte
	for {
		if r.stopped {
			break
		}
		if r.paused {
			<-r.pause
		}

		l, err := buf.ReadBytes('\n')
		line = append(line, l...)

		// 读取到文件结尾, 等待
		if err == io.EOF {
			<-time.After(time.Second)
			continue
		}
		if err != nil {
			return err
		}

		r.Channel <- string(line)
		r.offsets += len(line)
		r.lines++
		line = []byte{}
	}

	close(r.pause)
	close(r.Channel)
	return nil
}

// NewReader 创建新的Reader
func NewReader(isGzip bool, path string, offsets int, lines int) *Reader {
	r := &Reader{
		gzip:    isGzip,
		path:    path,
		offsets: offsets,
		lines:   lines,
		pause:   make(chan bool),
		Channel: make(chan string),
	}

	return r
}
