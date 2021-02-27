package writer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mholt/archiver"
	"github.com/projectdiscovery/gologger/levels"
)

func init() {
	// Set default dir to current directory + /log
	if dir, err := os.Getwd(); err == nil {
		DefaultFileWithRotationOptions.Location = path.Join(dir, "logs")
	}

	DefaultFileWithRotationOptions.RotationInterval = time.Hour

	// Current logfile name is "processname.log"
	DefaultFileWithRotationOptions.FileName = fmt.Sprintf("%s.log", filepath.Base(os.Args[0]))
}

// FileWithRotation is a concurrent output writer to a file with rotation.
type FileWithRotation struct {
	options *FileWithRotationOptions
	mutex   *sync.Mutex
	logFile *os.File
}

type FileWithRotationOptions struct {
	Location         string
	Rotate           bool
	RotationInterval time.Duration
	FileName         string
	Compress         bool
}

var DefaultFileWithRotationOptions FileWithRotationOptions

// NewFileWithRotation returns a new file concurrent log writer.
func NewFileWithRotation(options *FileWithRotationOptions) (*FileWithRotation, error) {
	fwr := &FileWithRotation{
		options: options,
		mutex:   &sync.Mutex{},
	}
	// set log rotator monitor
	if fwr.options.Rotate {
		go scheduler(time.NewTicker(options.RotationInterval), fwr.checkAndRotate)
	}

	err := os.MkdirAll(fwr.options.Location, 655)
	if err != nil {
		return nil, err
	}

	err = fwr.newLogger()
	if err != nil {
		return nil, err
	}

	return fwr, nil
}

// WriteString writes an output to the underlying file
func (w *FileWithRotation) Write(data []byte, level levels.Level) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	switch level {
	case levels.LevelSilent:
		w.logFile.Write(data)
		w.logFile.Write([]byte("\n"))
	default:
		w.logFile.Write(data)
		w.logFile.Write([]byte("\n"))
	}
}

func (w *FileWithRotation) checkAndRotate() {
	// extract time from filename
	ts := strings.TrimSuffix(strings.TrimPrefix(w.options.FileName, filepath.Base(os.Args[0])+"-"), ".log")
	t, err := time.Parse(backupTimeFormat, ts)
	if err != nil {
		return
	}

	// if current day is different from current one of log rotate
	if !dateEqual(t, time.Now()) {
		w.Close()
		// start asyncronous rotation
		if w.options.Compress {
			w.compressLogs()
		}
		w.newLogger()
	}
}

// Close and flushes the logger
func (w *FileWithRotation) Close() {
	w.logFile.Sync()
	w.logFile.Close()
}

func (w *FileWithRotation) newLogger() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// update file name if necessary
	if w.options.Rotate {
		w.buildRotateName()
	}

	logFile, err := w.CreateFile(w.options.FileName)
	if err != nil {
		return err
	}
	w.logFile = logFile

	return nil
}

func (w *FileWithRotation) buildRotateName() {
	timestamp := time.Now().Format(backupTimeFormat)
	fullPath := path.Join(w.options.Location, fmt.Sprintf("%s-%s.log", filepath.Base(os.Args[0]), timestamp))
	w.options.FileName = fullPath
}

func (w *FileWithRotation) CreateFile(filename string) (*os.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		f, err = os.Create(filename)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

func (w *FileWithRotation) compressLogs() {
	// snapshot current filename log
	filename := w.options.FileName
	// start asyncronous compressing
	go func(filename string) {
		err := archiver.CompressFile(filename, filename+".gz")
		if err == nil {
			// remove the original file
			os.RemoveAll(filename)
		}
	}(filename)
}

const (
	backupTimeFormat = "2006-01-02"
)

func scheduler(tick *time.Ticker, f func()) {
	for range tick.C {
		f()
	}
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func dateHourEqual(date1, date2 time.Time) bool {
	return dateEqual(date1, date2) && date1.Hour() == date2.Hour()
}
