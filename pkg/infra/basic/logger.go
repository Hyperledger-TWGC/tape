package basic

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func LogEvent(logger *log.Logger, txid, event string) {
	now := time.Now()
	time_str := fmt.Sprintf("yyyy-mm-dd HH:mm:ss: ", now.Format("2006-01-02 15:04:05.000000000"))
	logger.Debugf("For txid %s, event %s at %s", txid, event, time_str)
}
