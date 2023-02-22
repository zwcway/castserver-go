package decoder

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

const (
	localFile  = 1
	remoteFile = 2

	processClosed     = -1
	processDownloaded = -2
)

type FileStream struct {
	ctx     utils.Context
	log     *zap.Logger
	fp      *os.File
	rpath   string
	t       int
	rpos    int64
	wpos    int64
	size    int64
	process chan int64
	isDone  bool
	rw      sync.RWMutex
}

func (s *FileStream) Read(p []byte) (n int, err error) {
	if s.t == localFile {
		return s.fp.Read(p)
	}

	if s.isDone {
		n, err = s.fp.ReadAt(p, s.rpos)
		s.rpos += int64(n)
		return
	}

	defer func() {
		if errs := recover(); errs != nil {
			err = fmt.Errorf("read err :%s", errs)
		}
		s.rpos += int64(n)

		s.rw.Unlock()
	}()

	bufLen := len(p)

	var size int64
	for {
		// channel可能已满，但是不影响结果
		size = <-s.process // 正数文件总大小 或 负数表示状态
		s.rw.Lock()        // 立即获取读写锁

		if size == processClosed || s.isDone || s.wpos >= s.rpos+int64(bufLen) {
			// 有足够的内容可读取
			return s.fp.ReadAt(p, s.rpos)
		}

		s.rw.Unlock() // 等待下一次下载
	}
}

func (s *FileStream) Seek___(offset int64, whence int) (int64, error) {
	if s.t == localFile {
		return s.fp.Seek(offset, whence)
	}

	switch whence {
	case 0:
		s.rpos = offset
	case 1:
		s.rpos += offset
	case 2:
		if !s.isDone {
			return s.rpos, fmt.Errorf("file is downloading")
		}
		s.rpos = s.size - offset
	}

	return s.rpos, nil
}

func (s *FileStream) ReadAndRest(p []byte) (n int, err error) {
	n, err = s.Read(p)
	s.rpos = 0
	return
}

func (s *FileStream) Close() error {
	if s == nil {
		return nil
	}
	s.rpos = 0
	s.size = 0

	if s.process != nil {
		s.isDone = true // 防止 Read 方法进入循环
		s.process <- processClosed
		close(s.process)
		s.process = nil
	}

	if s.fp != nil {
		if s.t == remoteFile {
			os.Remove(s.fp.Name())
		}
		err := s.fp.Close()
		s.fp = nil
		return err
	}
	return nil
}

func (s *FileStream) downloadRoutine(resp *http.Response) {
	defer resp.Body.Close()
	src := resp.Body.(io.Reader)
	dst := s.fp
	s.wpos = 0
	isUnknownSize := s.size == 0
	s.isDone = false
	s.log.Info("begin download file", zap.String("local", s.fp.Name()), zap.String("url", s.rpath))
	// 参考 io.Copy
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	buf := make([]byte, size)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			s.rw.Lock()
			nw, ew := dst.WriteAt(buf[0:nr], s.wpos)
			s.rw.Unlock()
			if nw < 0 || nr < nw {
				break
			}
			if ew != nil {
				break
			}
			s.wpos += int64(nw)
			if nr != nw {
				break
			}

			if isUnknownSize {
				s.size = s.wpos
			}
			if len(s.process) < cap(s.process) { // 忽略 channel 已满，剩下的足够正常运行
				s.process <- s.wpos
			}
		}
		if er != nil {
			break
		}
	}
	s.log.Info("download file complete", zap.String("local", s.fp.Name()), zap.String("url", s.rpath), zap.Int64("size", s.wpos))
	s.isDone = true
	s.process <- processDownloaded
}

func (s *FileStream) readHttpFile(urlPath string) (err error) {
	var u *url.URL
	s.rpath = urlPath
	u, err = url.Parse(urlPath)
	if err != nil {
		return err
	}
	if u.Port() == "" {
		if strings.Contains(u.Host, ":") {
			// ipv6地址
			u.Host = "[" + u.Host + "]:80"
		} else {
			u.Host += ":80"
		}
	}
	cli := http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	tmpPath := config.TemporayFile(urlPath)
	if tmpPath == "" {
		return fmt.Errorf("can not read temp dir")
	}

	s.fp, err = os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModeAppend|os.ModePerm)
	if err != nil {
		return fmt.Errorf("can not open temp file '%s' %s", tmpPath, err.Error())
	}

	headresp, err := cli.Head(urlPath)
	if err != nil {
		return err
	}
	defer headresp.Body.Close()

	s.size, err = strconv.ParseInt(headresp.Header.Get("Content-Length"), 0, 64)
	if err != nil {
		s.size = 0
	}
	s.log.Info("need download", zap.String("local", s.fp.Name()), zap.String("url", s.rpath), zap.Int64("size", s.size))

	resp, err := cli.Get(urlPath)
	if err != nil {
		return err
	}

	s.process = make(chan int64, 10)

	go s.downloadRoutine(resp)

	return err
}

func (s *FileStream) OpenFile(oriPath string) (err error) {
	s.t = localFile
	s.size = 0

	if strings.HasPrefix(strings.ToLower(oriPath), "http") {
		s.t = remoteFile

		return s.readHttpFile(oriPath)
	} else {
		fi, err := os.Stat(oriPath)
		if err != nil {
			return err
		}
		s.size = fi.Size()
	}

	s.fp, err = os.OpenFile(oriPath, os.O_RDONLY, 0)

	return err
}

func NewFileStream(ctx utils.Context) *FileStream {
	s := &FileStream{
		ctx:    ctx,
		log:    ctx.Logger("filestream"),
		rpos:   0,
		wpos:   0,
		isDone: false,
	}
	return s
}
