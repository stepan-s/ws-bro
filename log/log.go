package log

import (
	"io"
	stdLog "log"
)

// Отключение логгирования
const NONE = 0

// Система находится в неработоспособном состоянии
const EMERGENCY = 1

// Требуется немедленное реагирование (например, база не доступна)
const ALERT = 2

// Критическая ошибка (что-то не работает, требует исправления)
const CRITICAL = 3

// Ошибка (рантайм, не требующая немедленных действий)
const ERROR = 4

// Предупреждение (необычная ситуация, но не ошибка)
const WARNING = 5

// Важная информация
const NOTICE = 6

// Обычная информация
const INFO = 7

// Детализированная отладочная информация
const DEBUG = 8

const sEMERGENCY = "EMERGENCY"
const sALERT = "ALERT"
const sCRITICAL = "CRITICAL"
const sERROR = "ERROR"
const sWARNING = "WARNING"
const sNOTICE = "NOTICE"
const sINFO = "INFO"
const sDEBUG = "DEBUG"

var log *Logger

type Logger struct {
	level  uint8
	logger *stdLog.Logger
}

func Init(w io.Writer, level uint8) {
	log = new(Logger)
	log.level = level
	log.logger = stdLog.New(w, "", stdLog.LstdFlags)
}

func (logger *Logger) write(level uint8, message *string, v ...interface{}) {
	var levelCaption string
	switch level {
	case NONE:
		return
	case EMERGENCY:
		levelCaption = sEMERGENCY
	case ALERT:
		levelCaption = sALERT
	case CRITICAL:
		levelCaption = sCRITICAL
	case ERROR:
		levelCaption = sERROR
	case WARNING:
		levelCaption = sWARNING
	case NOTICE:
		levelCaption = sNOTICE
	case INFO:
		levelCaption = sINFO
	case DEBUG:
		levelCaption = sDEBUG
	}
	logger.logger.Printf("["+levelCaption+"] "+*message, v...)
}

func Emergency(message string, v ...interface{}) {
	log.write(EMERGENCY, &message, v...)
}

func Alert(message string, v ...interface{}) {
	log.write(ALERT, &message, v...)
}

func Critical(message string, v ...interface{}) {
	log.write(CRITICAL, &message, v...)
}

func Error(message string, v ...interface{}) {
	log.write(ERROR, &message, v...)
}

func Warning(message string, v ...interface{}) {
	log.write(WARNING, &message, v...)
}

func Notice(message string, v ...interface{}) {
	log.write(NOTICE, &message, v...)
}

func Info(message string, v ...interface{}) {
	log.write(INFO, &message, v...)
}

func Debug(message string, v ...interface{}) {
	log.write(DEBUG, &message, v...)
}
