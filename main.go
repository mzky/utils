package main

import (
	"utils/log"
)

func main() {

	logger, err := log.New("sentry", "debug", "")
	if err != nil {
		log.Fatal("%s", err)
	}
	log.Export(logger)

	log.Info("info")
	log.Debug("debug")
	log.Warn("warn")
	log.Error("error")
	log.ReceiveMsg("ReceiveMsg")
	log.FightValueChange("FightValueChange")
	log.SendMsg("SendMsg")
	log.Recover("Recover")
	log.Fatal("fatal")
}
