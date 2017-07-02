package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

type ApiConf struct {
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

var (
	IncomingUrl string = "slackのwebhookURL"
)

type Slack struct {
	Text       string `json:"text"`       //投稿内容
	Username   string `json:"username"`   //投稿者名 or Bot名（存在しなくてOK）
	Icon_emoji string `json:"icon_emoji"` //アイコン絵文字
	Icon_url   string `json:"icon_url"`   //アイコンURL（icon_emojiが存在する場合は、適応されない）
	Channel    string `json:"channel"`    //#部屋名
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	var apiConf ApiConf
	{
		apiConfPath := flag.String("conf", "config.json", "API Config File")
		flag.Parse()
		data, err_file := ioutil.ReadFile(*apiConfPath)
		check(err_file)
		err_json := json.Unmarshal(data, &apiConf)
		check(err_json)
	}
	anaconda.SetConsumerKey(apiConf.ConsumerKey)
	anaconda.SetConsumerSecret(apiConf.ConsumerSecret)
	api := anaconda.NewTwitterApi(apiConf.AccessToken, apiConf.AccessTokenSecret)

	twitterStream := api.UserStream(nil)
	for {
		x := <-twitterStream.C
		switch tweet := x.(type) {
		case anaconda.EventTweet:
			var mediaURL string
			var tweetText string
			if tweet.Event.Event != "favorite" {
				continue
			}
			if tweet.TargetObject.ExtendedEntities.Media == nil {
				continue
			}
			//自分がいいねされても反応するからここで防止
			if tweet.TargetObject.User.ScreenName == "＠抜き自分のスクリーンネーム" {
				continue
			}
			for _, url := range tweet.TargetObject.ExtendedEntities.Media {
				mediaURL = mediaURL + url.Media_url + "\n"
			}
			tweetText = tweet.TargetObject.Text + "\n"
			postText := tweetText + mediaURL
			params, _ := json.Marshal(Slack{
				postText,
				"",
				"",
				"",
				"投げるslackチャンネル名"})
			resp, err := http.PostForm(
				IncomingUrl,
				url.Values{"payload": {string(params)}},
			)
			if err != nil {
				fmt.Print(err)
			}

			body, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			fmt.Println("posted slack: " + string(body))
		default:

		}
	}
}
