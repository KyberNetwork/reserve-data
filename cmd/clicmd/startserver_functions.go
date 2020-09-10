package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/common/archive"
)

const (
	logFileName = "core.log"
)

var (
	errorLevel = "error"
	warnLevel  = "warn"
	fatalLevel = "fatal"
	infoLevel  = "info"
)

func backupLog(arch archive.Archive, cronTimeExpression string, fileNameRegrexPattern string) {
	c := cron.New()
	l := zap.S()
	err := c.AddFunc(cronTimeExpression, func() {
		files, rErr := ioutil.ReadDir(logDir)
		if rErr != nil {
			l.Warnw("BACKUPLOG ERROR: Can not view log folder", "err", rErr)
			return
		}
		for _, file := range files {
			matched, err := regexp.MatchString(fileNameRegrexPattern, file.Name())
			if (!file.IsDir()) && (matched) && (err == nil) {
				err := arch.UploadFile(arch.GetLogBucketName(), remoteLogPath, filepath.Join(logDir, file.Name()))
				if err != nil {
					l.Warnw("BACKUPLOG ERROR: Can not upload Log file", "err", err)
				} else {
					var err error
					var ok bool
					if file.Name() != logFileName {
						ok, err = arch.CheckFileIntergrity(arch.GetLogBucketName(), remoteLogPath, filepath.Join(logDir, file.Name()))
						if !ok || (err != nil) {
							l.Warnw("BACKUPLOG ERROR: File intergrity is corrupted")
						}
						err = os.Remove(filepath.Join(logDir, file.Name()))
					}
					if err != nil {
						l.Warnw("BACKUPLOG ERROR: Cannot remove local log file", "err", err)
					} else {
						l.Infof("BACKUPLOG Log backup: backup file %s successfully", file.Name())
					}
				}
			}
		}
	})
	if err != nil {
		l.Warnw("BACKUPLOG Cannot rotate log", "err", err)
	}
	c.Start()
}
func newLogger(writeTo io.Writer) (*zap.Logger, error) {
	var encoder zapcore.Encoder
	switch zapMode {
	case "prod":
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	default:
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	c := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(writeTo), zap.DebugLevel))
	return c, nil
}

type syncer interface {
	Sync() error
}

func newFlusher(s syncer) func() {
	return func() {
		// ignore the error as the sync function will always fail in Linux
		// https://github.com/uber-go/zap/issues/370
		_ = s.Sync()
	}
}

func newSugaredLogger(w io.Writer) (*zap.SugaredLogger, func(), error) {
	logger, err := newLogger(w)
	if err != nil {
		return nil, nil, err
	}
	// init sentry if flag dsn exists
	if len(sentryDSN) != 0 {
		sentryClient, err := sentry.NewClient(sentry.ClientOptions{
			Dsn: sentryDSN,
		})
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to init sentry client")
		}

		cfg := zapsentry.Configuration{
			DisableStacktrace: false,
		}
		switch sentryLevel {
		case infoLevel:
			cfg.Level = zapcore.InfoLevel
		case warnLevel:
			cfg.Level = zapcore.WarnLevel
		case errorLevel:
			cfg.Level = zapcore.ErrorLevel
		case fatalLevel:
			cfg.Level = zapcore.FatalLevel
		default:
			return nil, nil, errors.Errorf("invalid log level %v", sentryDSN)
		}

		z, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(sentryClient))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to init zap sentry")
		}
		// attach to logger core
		logger = zapsentry.AttachCoreToLogger(z, logger)
	}
	sugar := logger.Sugar()
	return sugar, newFlusher(logger), nil
}

//set config log: Write log into a predefined file, and rotate log daily
//if stdoutLog is set, the log is also printed on stdout.
func configLog(stdoutLog bool) io.Writer {
	logger := &lumberjack.Logger{
		Filename: filepath.Join(logDir, logFileName),
		// MaxSize:  1, // megabytes
		MaxBackups: 0,
		MaxAge:     0, //days
		// Compress:   true, // disabled by default
	}
	var mw io.Writer
	if stdoutLog {
		mw = io.MultiWriter(os.Stdout, logger)
	} else {
		mw = io.MultiWriter(logger)
	}
	log.SetOutput(mw)

	c := cron.New()
	err := c.AddFunc("@daily", func() {
		if lErr := logger.Rotate(); lErr != nil {
			panic(fmt.Sprintf("error rotate log: %v", lErr))
		}
	})
	if err != nil {
		panic(fmt.Sprintf("error add log cron daily: %v", err))
	}
	c.Start()
	return mw
}

// CreateBlockchain create new blockchain instance
func CreateBlockchain(config *configuration.AppState) (bc *blockchain.Blockchain, err error) {
	bc, err = blockchain.NewBlockchain(config.Blockchain, config.Setting)

	if err != nil {
		panic(err)
	}

	// load all listed token indices instead of only configs tokens
	err = bc.LoadAndSetTokenIndices()
	if err != nil {
		log.Panicf("Can't load and set token indices: %s", err)
	}
	return
}
