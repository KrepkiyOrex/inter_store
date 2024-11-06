package utils

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// определяем цвета для разных уровней логирования
const (
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	reset   = "\033[0m"
)

// CustomFormatter для настройки формата логов
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Форматирование времени
	timeFormat := entry.Time.Format("2006-01-02 15:04:05 -07:00") // Добавлен пробел между датой и временем

	// Определение цвета в зависимости от уровня логирования
	var levelColor string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = blue
	case logrus.InfoLevel:
		levelColor = green
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel:
		levelColor = red
	case logrus.FatalLevel:
		levelColor = red
	default:
		levelColor = reset
	}

	// форматирование строки с учетом цвета
	return []byte(fmt.Sprintf("%s [%s%s%s] %s\n", timeFormat, levelColor, entry.Level, reset, entry.Message)), nil
}

// InitLogrus инициализирует логгер Logrus
func InitLogrus(level logrus.Level, jsonFormat, enableColors bool) {
	// устанавливаем уровень логирования
	logrus.SetLevel(level)

	if jsonFormat {
		// JSON формат без цветных логов
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		// устанавливаем кастомный форматер
		logrus.SetFormatter(&CustomFormatter{})
	}

	// устанавливаем вывод в стандартный поток
	logrus.SetOutput(os.Stdout)
}
