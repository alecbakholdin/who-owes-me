package venmo

import (
	"encoding/json"
	"fmt"
	"strings"
)

type UserPageNextData struct {
	Props struct {
		PageProps struct {
			OtherUser UserData
		}
	}
}

type UserData struct {
	DisplayName       string
	Id                string `json:"id"`
	Username          string
	FirstName         string
	LastName          string
	ProfilePictureUrl string
}

func (v *Client) GetUser(venmo string) (*UserData, error) {
	formatted := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(venmo), "@"))
	page, err := v.initPage()
	if err != nil {
		return nil, err
	}
	r, err := page.Goto(fmt.Sprintf("https://account.venmo.com/u/%s", formatted))
	if err != nil {
		return nil, err
	}
	if r.Status() != 200 {
		return nil, fmt.Errorf("non-200 status code finding %s %d", venmo, r.Status())
	}

	bytes, err := page.Locator("#__NEXT_DATA__").InnerText()
	if err != nil {
		return nil, err
	}

	data := UserPageNextData{}
	if err = json.Unmarshal([]byte(bytes), &data); err != nil {
		return nil, err
	}
	return &data.Props.PageProps.OtherUser, nil
}
