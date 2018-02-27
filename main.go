package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	log "github.com/sirupsen/logrus"
)

var CONSUMER_KEY = os.Getenv("TWITTER_CONSUMER_KEY")
var CONSUMER_SECRET = os.Getenv("TWITTER_CONSUMER_SECRET")
var ACCESS_TOKEN = os.Getenv("TWITTER_OAUTH_TOKEN")
var ACCESS_TOKEN_SECRET = os.Getenv("TWITTER_OAUTH_TOKEN_SECRET")

type logFormatter struct{}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

func main() {
	log.SetLevel(log.DebugLevel)

	log.SetFormatter(new(logFormatter))

	if len(os.Args) != 3 {
		log.Errorf("usage: tweepfilter [screen_name] [filter]")
		return
	}

	sourceUser := os.Args[1]
	filter := strings.ToLower(os.Args[2])

	if CONSUMER_KEY == "" || CONSUMER_SECRET == "" || ACCESS_TOKEN == "" || ACCESS_TOKEN_SECRET == "" {
		log.Errorf("env vars not set")
		return
	}

	anaconda.SetConsumerKey(CONSUMER_KEY)
	anaconda.SetConsumerSecret(CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(ACCESS_TOKEN, ACCESS_TOKEN_SECRET)
	defer api.Close()

	rlsr, err := api.GetRateLimits([]string{"friends,friendships"})
	if err != nil {
		log.Errorf("GetRateLimits failed with: %+v", err)
		return
	}
	log.Debugf("RateLimitStatusResponse %+v", rlsr)

	v := url.Values{
		"screen_name":           []string{sourceUser},
		"count":                 []string{"200"},
		"exclude_replies":       []string{"true"},
		"include_user_entities": []string{"true"},
		"skip_status":           []string{"true"},
	}

	cursor := "-1"

	for {
		log.Debugf("CURSOR: %+v", cursor)

		v.Set("cursor", cursor)

		c, err := api.GetFriendsList(v)
		if err != nil {
			log.Errorf("GetFriendsList failed with: %+v", err)
			return
		}

		cursor = c.Next_cursor_str

		for _, user := range c.Users {
			userStr := strings.ToLower(fmt.Sprintf("%v", user))
			if strings.Contains(userStr, filter) {
				log.Debugf("%+v", user)
				log.Infof("%s", user.ScreenName)

				if strings.Contains(strings.ToLower(user.Description), filter) {
					log.Infof("  %s", user.Description)
				}

				for _, url := range user.Entities.Url.Urls {
					if strings.Contains(strings.ToLower(url.Expanded_url), filter) {
						log.Infof("  A: %s", url.Expanded_url)
					}
				}
				for _, url := range user.Entities.Urls {
					if strings.Contains(strings.ToLower(url.Expanded_url), filter) {
						log.Infof("  B: %s", url.Expanded_url)
					}
				}
			}
		}
	}
}
