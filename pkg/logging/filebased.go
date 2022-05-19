package logging

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var (
	fileStringFormat = func() (ret string) {
		ret = time.RFC822
		ret = strings.ReplaceAll(ret, " ", "-")
		ret = strings.ReplaceAll(ret, ":", "-")
		return
	}()
)

type ThreadSafeLoggingSystem struct {
	mu sync.Mutex

	FileLoggingEnabled bool
	FileLocation       string
	FileName           string
	File               *os.File
	RunStart           time.Time
}

func (ls *ThreadSafeLoggingSystem) OpenFile() (err error) {
	var wd string
	if wd, err = os.Getwd(); err != nil {
		return
	}

	ls.FileName = fmt.Sprintf(fileStringFormat+".log", ls.RunStart)

	var fp string
	if strings.Contains(ls.FileLocation, wd) {
		fp = path.Join(ls.FileLocation, ls.FileName)
	} else {
		fp = path.Join(wd, ls.FileLocation, ls.FileName)
	}

	if ls.File, err = os.Open(fp); err != nil {
		return
	}

	return
}

var (
	logStringFmt = fmt.Sprintf("[%v]: %%v\n", time.RFC1123Z)
)

func (ls *ThreadSafeLoggingSystem) LogString(msg string) string {
	return fmt.Sprintf(, time.Now(), msg)
}

func (ls *ThreadSafeLoggingSystem) Write(msg string) (err error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	_, err = ls.File.WriteString(msg)
	return
}

func (ls *ThreadSafeLoggingSystem) Writef(msg string, args ...interface{}) (err error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	_, err = ls.File.WriteString(
		fmt.Sprintf(msg, args...),
	)
	return
}

func NewFileLoggingSystem() *ThreadSafeLoggingSystem {
	return &ThreadSafeLoggingSystem{
		mu:       sync.Mutex{},
		RunStart: time.Now(),
	}
}
