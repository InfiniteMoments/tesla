package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"net/url"
)

func main() {
	anaconda.SetConsumerKey("OoNh7FzVLSs2Yte74jt7M0Cq6")
	anaconda.SetConsumerSecret("tA3MRJjDXUw6hgeRiUuroUbKr9M4pmz9ibIgoASevynTlOPLEr")
	api := anaconda.NewTwitterApi(
		"2918216473-EPfbdcB86hcq2ukFzVba62LOhc2aRWZC64UTOrd", "agGJ73g193mSHRsF5EtzDwNb3XkB4UNa8N0kANb8a6iQ9")

	v := url.Values{}
	s := api.PublicStreamSample(v)

	for t := range s.C {
		switch v := t.(type) {
		case anaconda.Tweet:
			fmt.Printf("%-15s: %s\n", v.User.ScreenName, v.Text)
		case anaconda.EventTweet:
			switch v.Event.Event {
			case "favorite":
				sn := v.Source.ScreenName
				tw := v.TargetObject.Text
				fmt.Printf("★ Favorited by %-15s: %s\n", sn, tw)
			case "unfavorite":
				sn := v.Source.ScreenName
				tw := v.TargetObject.Text
				fmt.Printf("★ UnFavorited by %-15s: %s\n", sn, tw)
			}
		}
	}

	//// TESTING throttling
	//api.SetDelay(10 * time.Second)

	// searchResult, err := api.GetSearch("#isitjustme", nil)
	// for _, tweet := range searchResult.Statuses {
	// 	fmt.Println(tweet.Text)
	// }

	// if err != nil {
	// 	fmt.Println(err)
	// }

}
