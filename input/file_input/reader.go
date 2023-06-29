package file_input

import (
	"bufio"
	"io"
	"os"
	"time"
)

type Reader struct {
	path    string
	stopped bool
	paused  bool
	pause   chan bool
	offsets int
	lines   int
	Channel chan string
}

// 停止
func (r *Reader) Stop() {
	r.stopped = true
}

// 暂停
func (r *Reader) Pause() {
	r.paused = true
}

// 继续
func (r *Reader) Continue() {
	r.paused = false
	r.pause <- true
}

// 当前行数
func (r *Reader) Lines() int {
	return r.lines
}

// 偏移量
func (r *Reader) Offsets() int {
	return r.offsets
}

// 开始监听
func (r *Reader) Watch() error {
	f, err := os.Open(r.path)

	if err != nil {
		return err
	}

	defer f.Close()

	buf := bufio.NewReader(f)
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

// 创建新的Reader
func NewReader(path string, offsets int, lines int) *Reader {
	r := &Reader{
		path:    path,
		offsets: offsets,
		lines:   lines,
		pause:   make(chan bool),
		Channel: make(chan string),
	}

	return r
}
