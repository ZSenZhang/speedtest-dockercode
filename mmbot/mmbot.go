package mmbot

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	PanicIcon   = "https://crowdtrace.caida.org/assets/images/warning.svg"
	DefaultIcon = "https://crowdtrace.caida.org/assets/images/avatars/earth.svg"
)

type MMPayload struct {
	Username string `json:"username"`
	Icon     string `json:"icon_url"`
	Text     string `json:"text"`
}

type MMBot struct {
	Webhook  string `json:"webhook"`
	Username string `json:"name"`
	Icon     string `json:"icon"`
	Quiet    bool
}

func NewMMBot(cfgfile string) *MMBot {
	mm := &MMBot{}
	if len(cfgfile) > 0 {
		cfile, err := os.Open(cfgfile)
		defer cfile.Close()
		if err != nil {
			log.Fatal(err)
		}
		mm.Quiet = false
		decoder := json.NewDecoder(cfile)
		//	cfg := &MMCfg{}
		err = decoder.Decode(mm)
		if err != nil {
			log.Fatal(err)
		}
		if len(mm.Icon) == 0 {
			mm.Icon = DefaultIcon
		}
	} else {
		mm.Quiet = true
	}
	return mm
}

func (mm *MMBot) Test() error {
	if mm.Quiet {
		log.Println("Quiet mode enabled")
		return nil
	}
	return mm.SendMsg("This is a test", DefaultIcon)
}
func (mm *MMBot) SendPanic(msg string) error {
	if !mm.Quiet {
		return mm.SendMsg(msg, PanicIcon)
	}
	return nil
}
func (mm *MMBot) SendInfo(msg string) error {
	if !mm.Quiet {
		return mm.SendMsg(msg, mm.Icon)
	}
	return nil
}

func (mm *MMBot) SendMsg(msg string, icon string) error {
	payload := MMPayload{}
	payload.Username = mm.Username
	payload.Icon = icon
	payload.Text = msg
	mpayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	urlv := url.Values{}
	urlv.Add("payload", string(mpayload))
	_, err = http.PostForm(mm.Webhook, urlv)
	return err
}
