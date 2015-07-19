package controller

import "log"

type twitterStreamConfig struct {
	searchQuery string
	stopChannel chan bool
}

type streamConfig struct {
	twitterStreamConfig
}

var currentlyStreaming []streamConfig

func StartStream(searchQuery string) {
	var stopChannel = make(chan bool)
	streamObject := streamConfig{}
	streamObject.twitterStreamConfig = twitterStreamConfig{searchQuery: searchQuery, stopChannel: stopChannel}
	currentlyStreaming = append(currentlyStreaming, streamObject)
	log.Println(currentlyStreaming)
}

func StopStream(searchQuery string) {

}
