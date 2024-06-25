package venmo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/playwright-community/playwright-go"
)

type Client struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	page    playwright.Page
}

func NewClient() *Client {
	return &Client{}
}

func LoadClient(filepath string) (*Client, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error opening file"), err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error reading file"), err)
	}

	cookies := []playwright.OptionalCookie{}
	if err := json.Unmarshal(bytes, &cookies); err != nil {
		return nil, errors.Join(fmt.Errorf("error unmarshaling json"), err)
	}
	client := NewClient()
	page, err := client.initPage()
	if err != nil {
		return nil, err
	}
	if err := page.Context().AddCookies(cookies); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) StoreToFile(filepath string) error {
	if c.page == nil {
		return nil
	}
	cookies, err := c.page.Context().Cookies()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(cookies)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(bytes)
	return err
}

func (v *Client) initPage() (playwright.Page, error) {
	if v.page != nil {
		return v.page, nil
	}

	var err error
	if v.pw == nil {
		fmt.Println("Running playwright")
		v.pw, err = playwright.Run()
	}

	if err != nil {
		v.pw = nil
		return nil, err
	}

	if v.browser == nil {
		fmt.Println("Launching browser")
		v.browser, err = v.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
		})
	}
	if err != nil {
		return nil, err
	}

	fmt.Println("Creating page")
	v.page, err = v.browser.NewPage()
	if err != nil {
		return nil, err
	}

	return v.page, err
}

func (v *Client) get(url string) ([]byte, error) {
	if page, err := v.initPage(); err != nil {
		return nil, err
	} else if response, err := page.Request().Get(url); err != nil {
		return nil, err
	} else if bytes, err := response.Body(); err != nil {
		return nil, err
	} else {
		return bytes, nil
	}
}

func (v *Client) post(url string, body any) ([]byte, error) {
	return v.postCsrf(url, body, "")
}

func (v *Client) postCsrfUrl(url string, body any, csrfUrl string) ([]byte, error) {
	csrf, err := v.csrf(csrfUrl)
	if err != nil {
		return nil, err
	}
	return v.postCsrf(url, body, csrf)
}

func (v *Client) postCsrf(url string, body any, csrf string) ([]byte, error) {
	fmt.Println("csrf", csrf)
	opts := playwright.APIRequestContextPostOptions{
		Data: body,
	}
	if page, err := v.initPage(); err != nil {
		return nil, err
	} else if response, err := page.Request().Post(url, opts); err != nil {
		return nil, err
	} else if response.Status() != 200 {
		return nil, fmt.Errorf("unexpectded status %d posting to %s", response.Status(), url)
	} else if bytes, err := response.Body(); err != nil {
		return nil, err
	} else {
		return bytes, nil
	}
}

type CsrfNextData struct {
	Props struct {
		PageProps struct {
			CsrfToken string
		}
	}
}

func (v *Client) csrf(pageUrl string) (string, error) {
	page, err := v.initPage()
	if err != nil {
		return "", err
	}
	_, err = page.Goto(pageUrl)
	if err != nil {
		return "", err
	}
	text, err := page.Locator("#__NEXT_DATA__").TextContent()
	if err != nil {
		return "", err
	}
	data := CsrfNextData{}
	err = json.Unmarshal([]byte(text), &data)
	if err != nil {
		return "", err
	}
	if data.Props.PageProps.CsrfToken == "" {
		return "", errors.New("empty csrf")
	}
	page.SetExtraHTTPHeaders(map[string]string{
		"Xsrf-Token": data.Props.PageProps.CsrfToken,
		"Csrf-Token": data.Props.PageProps.CsrfToken,
	})
	return data.Props.PageProps.CsrfToken, nil
}
