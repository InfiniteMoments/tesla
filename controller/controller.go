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
	for _, config := range currentlyStreaming {

		//Stop twitter stream
		if config.twitterStreamConfig.searchQuery == stopQuery {
			config.twitterStreamConfig.stopChannel <- stopQuery
			return
		}
	}
	log.Println("Record not found for search string:", stopQuery)
}
