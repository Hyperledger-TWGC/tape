package basic

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func LogEvent(logger *log.Logger, txid, event string) {
	now := time.Now()
	time_str := fmt.Sprintf(now.Format(time.RFC3339Nano))
	logger.Debugf("For txid %s, event %s at %s", txid, event, time_str)
}
