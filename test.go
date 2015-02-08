package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/spf13/viper"
	"net/url"
)

func initConfig() {
	viper.SetConfigName("moments_config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}
}

func main() {
	initConfig()

	anaconda.SetConsumerKey(viper.GetString("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(viper.GetString("CONSUMER_SECRET"))
	api := anaconda.NewTwitterApi(
		viper.GetString("ACCESS_TOKEN"), viper.GetString("ACCESS_SECRET"))

	v := url.Values{}
	v.Set("track", "#KRISHANDAY")
	s := api.PublicStreamFilter(v)

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
			fmt.Println(v)
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
