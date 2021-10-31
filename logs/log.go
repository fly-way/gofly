package logs

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// 日志标记
const (
	lDate         = 1 << iota                             // 日期 YYYY/MM/DD, 如: 2009/01/23
	lTime                                                 // 时间 hh:mm:ss, 如: 01:23:23
	lMicroseconds                                         // 微妙 hh:mm:ss.us, 如: 01:23:23.123123. 覆盖LTime.
	lLongFile                                             // 完整文件名和行号, 如: /a/b/c/d.go:23
	lShortFile                                            // 最红文件名和行号, 如: d.go:23. 覆盖 lLongFile
	lDayFile                                              // 日志文件带有时间后缀, 跨天会写入到新的文件
	lStdFlags     = lShortFile | lMicroseconds | lDayFile // 标准格式
)

type logger struct {
	mu       sync.Mutex // ensures atomic writes; protects the following fields
	flag     int        // 日志标记
	out      io.Writer  // 日志输出描述
	buf      []byte     // 输出缓冲区
	file     *os.File   // 日志文件
	fileDay  int        // 当前日志日期
	fileDir  string     // 日志文件目录
	fileName string     // 日志文件名称
	title    string     // 日志标识
	depth    int
}

// 新建日志文件
func new(flag int, title string, depth int) *logger {
	return &logger{out: os.Stderr, flag: flag, title: title, depth: depth, fileDay: -1}
}

// 设置日志文件输出路径
func (l *logger) setOutput(dir string, fileName string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.fileDir = dir
	l.fileName = fileName
}

// 日志输出
func (l *logger) output(s string) error {
	now := time.Now()
	var file string
	var line int

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.flag&lDayFile != 0 {
		if l.fileDay != now.Day() && l.openFile() {
			l.fileDay = now.Day()
		}
	} else {
		if l.file == nil {
			l.openFile()
		}
	}

	if l.flag&(lShortFile|lLongFile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(l.depth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)

	// TODO 临时输出
	fmt.Print(string(l.buf))
	return err
}

func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// 打开日志文件
func (l *logger) openFile() bool {
	if l.fileDir == "" || l.fileName == "" {
		return false
	}

	if l.file != nil {
		l.file.Close()
		l.file = nil
		l.out = os.Stderr
	}

	_, er := os.Stat(l.fileDir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(l.fileDir, 0775); err != nil {
			if os.IsPermission(err) {
				return false
			}
		}
	}

	var (
		buf     = make([]byte, 0, 32)
		y, m, d = time.Now().Date()
	)

	buf = append(buf, l.fileDir...)
	buf = append(buf, '/')
	buf = append(buf, l.fileName...)
	if l.flag&lDayFile != 0 {
		buf = append(buf, '_')
		itoa(&buf, y, 4)
		buf = append(buf, '_')
		itoa(&buf, int(m), 2)
		buf = append(buf, '_')
		itoa(&buf, d, 2)
	}
	buf = append(buf, ".log"...)

	var file *os.File
	if _, err := os.Stat(string(buf)); os.IsNotExist(err) {
		file, _ = os.OpenFile(string(buf), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	} else {
		file, _ = os.OpenFile(string(buf), os.O_APPEND|os.O_RDWR, 0644)
	}

	l.file = file
	l.out = file

	return true
}

// 格式化日志标头
func (l *logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
	if l.flag&(lDate|lTime|lMicroseconds) != 0 {
		if l.flag&lDate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(lTime|lMicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&lMicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}

	*buf = append(*buf, l.title...)
	if l.flag&(lShortFile|lLongFile) != 0 {
		if l.flag&lShortFile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}
