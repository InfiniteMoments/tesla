package controller

import (
	"log"

	"github.com/InfiniteMoments/tesla/twitter"
)

type twitterStreamConfig struct {
	searchQuery string
	stopChannel chan string
}

type streamConfig struct {
	twitterStreamConfig
}

var currentlyStreaming []streamConfig

func StartStream(searchQuery string) {
	var stopChannel = make(chan string)
	streamObject := streamConfig{}
	streamObject.twitterStreamConfig = twitterStreamConfig{searchQuery: searchQuery, stopChannel: stopChannel}
	currentlyStreaming = append(currentlyStreaming, streamObject)

	//Start twitter stream
	twitter.StartTwitterStream(searchQuery, stopChannel)
}

func StopStream(stopQuery string) {
	for i, config := range currentlyStreaming {

		//Stop twitter stream
		if config.twitterStreamConfig.searchQuery == stopQuery {
			config.twitterStreamConfig.stopChannel <- stopQuery
			currentlyStreaming = append(currentlyStreaming[:i], currentlyStreaming[i+1:]...)
			return
		}
	}
	log.Println("Record not found for search string:", stopQuery)
}

func StopAllStreams() {
	for _, config := range currentlyStreaming {
		config.twitterStreamConfig.stopChannel <- config.twitterStreamConfig.searchQuery
	}
}
