// Inspired by https://github.com/natefinch/lumberjack

package writer

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mholt/archiver"
	"github.com/projectdiscovery/gologger/levels"
	"gopkg.in/djherbis/times.v1"
)

func init() {
	// Set default dir to current directory + /logs
	if dir, err := os.Getwd(); err == nil {
		DefaultFileWithRotationOptions.Location = filepath.Join(dir, "logs")
	}

	// DefaultFileWithRotationOptions.rotationcheck = time.Duration(1 * time.Minute)
	DefaultFileWithRotationOptions.rotationcheck = time.Duration(5 * time.Second)

	// Current logfile name is "processname.log"
	DefaultFileWithRotationOptions.FileName = fmt.Sprintf("%s.log", filepath.Base(os.Args[0]))
	DefaultFileWithRotationOptions.BackupTimeFormat = "2006-01-02T15-04-05"
	DefaultFileWithRotationOptions.ArchiveFormat = "gz"
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
	rotationcheck    time.Duration
	RotationInterval time.Duration
	FileName         string
	Compress         bool
	MaxSize          int
	BackupTimeFormat string
	ArchiveFormat    string
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
		go scheduler(time.NewTicker(options.rotationcheck), fwr.checkAndRotate)
	}

	err := os.MkdirAll(fwr.options.Location, 0644)
	if err != nil {
		return nil, err
	}

	err = fwr.newLoggerSync()
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
	// check size
	currentFileSizeMb, err := w.logFile.Stat()
	if err != nil {
		return
	}

	filename := filepath.Join(w.options.Location, w.options.FileName)
	filebirthdate, err := getCreationTime(filename)
	if err != nil {
		return
	}

	filesizeCheck := w.options.MaxSize > 0 && currentFileSizeMb.Size() >= int64(w.options.MaxSize*1024*1024)
	filebirthdateCheck := w.options.RotationInterval > 0 && filebirthdate.Add(w.options.RotationInterval).Before(time.Now())

	// Rotate if:
	// - Size excedeed
	// - File max age excedeed
	if filesizeCheck || filebirthdateCheck {
		w.mutex.Lock()
		w.Close()
		w.compressLogs()
		w.newLogger()
		w.mutex.Unlock()
	}
}

// Close and flushes the logger
func (w *FileWithRotation) Close() {
	w.logFile.Sync()
	w.logFile.Close()
}

func (w *FileWithRotation) newLoggerSync() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.newLogger()
}

func (w *FileWithRotation) newLogger() (err error) {
	filename := filepath.Join(w.options.Location, w.options.FileName)
	logFile, err := w.CreateFile(filename)
	if err != nil {
		return err
	}
	w.logFile = logFile

	return nil
}

func (w *FileWithRotation) CreateFile(filename string) (*os.File, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (w *FileWithRotation) compressLogs() {
	// snapshot current filename log
	filename := filepath.Join(w.options.Location, w.options.FileName)
	fileExt := filepath.Ext(filename)
	filenameBase := strings.TrimSuffix(filename, fileExt)
	tmpFilename := filenameBase + "." + time.Now().Format(w.options.BackupTimeFormat) + fileExt
	os.Rename(filename, tmpFilename)

	if w.options.Compress {
		// start asyncronous compressing
		go func(filename string) {
			err := archiver.CompressFile(tmpFilename, filename+"."+w.options.ArchiveFormat)
			if err == nil {
				// remove the original file
				os.RemoveAll(tmpFilename)
			}
		}(tmpFilename)
	}
}

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

func getCreationTime(filename string) (*time.Time, error) {
	t, err := times.Stat(filename)
	if err != nil {
		return nil, err
	}

	if t.HasBirthTime() {
		birthTime := t.BirthTime()
		return &birthTime, nil
	}

	return nil, errors.New("No creation time")
}

func randStr(len int) string {
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)
	return str[:len]
}
