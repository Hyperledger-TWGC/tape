package basic

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func LogEvent(logger *log.Logger, txid, event string) {
	now := time.Now()
	year, month, day := now.Date()
	time_str := fmt.Sprintf("%d-%d-%d 00:00:00", year, month, day)
	logger.Debugf("For txid %s, event %s at %s", txid, event, time_str)
}
