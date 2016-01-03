package main

import (
	"encoding/json"
	"fmt"
	"github.com/tucnak/telebot"
	"github.com/yanple/vk_api"
	"log"
	"strconv"
	"time"
)

type ResponseUser struct {
	Response []struct {
		Uid       int64  `json:"uid"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"response"`
}

type ResponseWall struct {
	Response struct {
		Count int64 `json:"count"`
		Items []struct {
			ID          int64 `json:"id"`
			Attachments []struct {
				Doc struct {
					ID        int64  `json:"id"`
					AccessKey string `json:"access_key"`
					Date      int64  `json:"date"`
					Ext       string `json:"ext"`
					OwnerID   int64  `json:"owner_id"`
					Size      int64  `json:"size"`
					Title     string `json:"title"`
					URL       string `json:"url"`
				} `json:"doc"`
				Photo struct {
					ID        int64  `json:"id"`
					AccessKey string `json:"access_key"`
					AlbumID   int64  `json:"album_id"`
					Date      int64  `json:"date"`
					Height    int64  `json:"height"`
					OwnerID   int64  `json:"owner_id"`
					Photo130  string `json:"photo_130"`
					Photo604  string `json:"photo_604"`
					Photo75   string `json:"photo_75"`
					Photo807  string `json:"photo_807"`
					PostID    int64  `json:"post_id"`
					Text      string `json:"text"`
					UserID    int64  `json:"user_id"`
					Width     int64  `json:"width"`
				} `json:"photo"`
				Type string `json:"type"`
			} `json:"attachments, attachment"`
			CanDelete int64 `json:"can_delete"`
			Comments  struct {
				CanPost int64 `json:"can_post"`
				Count   int64 `json:"count"`
			} `json:"comments"`
			Date   int64 `json:"date"`
			FromID int64 `json:"from_id"`
			Likes  struct {
				CanLike    int64 `json:"can_like"`
				CanPublish int64 `json:"can_publish"`
				Count      int64 `json:"count"`
				UserLikes  int64 `json:"user_likes"`
			} `json:"likes"`
			OwnerID    int64 `json:"owner_id"`
			PostSource struct {
				Platform string `json:"platform"`
				Type     string `json:"type"`
			} `json:"post_source"`
			PostType string `json:"post_type"`
			Reposts  struct {
				Count        int64 `json:"count"`
				UserReposted int64 `json:"user_reposted"`
			} `json:"reposts"`
			Text string `json:"text"`
		} `json:"items"`
	} `json:"response"`
}

func post_from_bit_group(bot *telebot.Bot, message_chat telebot.Chat, stop chan telebot.Chat) {
	var api vk_api.Api
	api.AccessToken = ""
	api.UserId = 5173812

	params := make(map[string]string)
	params["domain"] = "bit_p3450"
	params["filter"] = "all"
	params["offset"] = "1"
	params["v"] = "5.42"

	strResp, err := api.Request("wall.get", params)
	if err != nil {
		fmt.Println("error in wall.get")
	}

	prev := ResponseWall{}

	if err := json.Unmarshal([]byte(strResp), &prev); err != nil {
		fmt.Println("error in decode json")
	}

	for {
		if stop_chat := <-stop; stop_chat == message_chat {
			//log.Println("Lets done")
			break
		}
		strResp, err := api.Request("wall.get", params)
		if err != nil {
			log.Println("error in wall.get")
		}
		curr := ResponseWall{}
		if err := json.Unmarshal([]byte(strResp), &curr); err != nil {
			log.Println("error in decode json")
		}
		//log.Println(curr.Response.Count)
		//log.Println(prev.Response.Count)
		if curr.Response.Count != prev.Response.Count {
			log.Println("found new message")
			message := "New message from "
			/* Recieve User name and last name if it possible */
			if curr.Response.Items[0].OwnerID != 0 {
				user_params := make(map[string]string)
				user_params["version"] = "5.42"
				user_params["user_ids"] = strconv.Itoa(int(curr.Response.Items[0].FromID))
				user_resp, err := api.Request("users.get", user_params)
				if err != nil {
					fmt.Println("error in wall.get")
				}
				user := ResponseUser{}
				if err := json.Unmarshal([]byte(user_resp), &user); err != nil {
					log.Println("error in decode user json")
				}
				message += user.Response[0].FirstName + " " + user.Response[0].LastName
			}
			message += "\n"
			message += curr.Response.Items[0].Text
			message += "\n"
			if curr.Response.Items[0].Attachments != nil {
				message += "Attachment \n"
				if curr.Response.Items[0].Attachments[0].Type == "doc" {
					message += "Doc Files, see it in group"
				}
				if curr.Response.Items[0].Attachments[0].Type == "photo" {
					message += "Photo"
				}
			}
			bot.SendMessage(message_chat, message, nil)
			prev = curr
		}
	}
}
func main() {
	bot, err := telebot.NewBot("")
	if err != nil {
		return
	}

	messages := make(chan telebot.Message)
	bot.Listen(messages, 1*time.Second)

	stop := make(chan telebot.Chat)

	go func() {
		for {
			time.Sleep(time.Second * 5)
			stop <- telebot.Chat{}
		}
	}()

	for message := range messages {
		if message.Text == "/help" {
			bot.SendMessage(message.Chat,
				"Hello, "+message.Sender.FirstName+"!\nThis bot is news notifier\nSend /start to start or /stop to stop", nil)
		}
		if message.Text == "/start" {
			go post_from_bit_group(bot, message.Chat, stop)
			bot.SendMessage(message.Chat, "Okey, starting notification", nil)
		}
		if message.Text == "/stop" {
			stop <- message.Chat
			bot.SendMessage(message.Chat, "Stopping -:_:-", nil)
		}
	}
}
