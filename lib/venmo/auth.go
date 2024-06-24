package venmo

import (
	"fmt"
)

func (v *Client) Login(username, password string) error {
	if loggedIn, err := v.IsLoggedIn(); err != nil {
		return err
	} else if loggedIn {
		return nil
	}
	page, err := v.initPage()
	if err != nil {
		return err
	}
	if _, err = page.Goto("https://id.venmo.com/signin#/lgn"); err != nil {
		return err
	}


	fmt.Println("Filling in email")
	email := page.Locator("input[type='email']")
	if populated, err := email.IsHidden(); err != nil {
		return err
	} else if populated {
		fmt.Println("email is already populated. Skipping filling in email")
	}	else if err := email.Click(); err != nil {
		return err
	} else if err := email.Fill(username); err != nil {
		return err
	} else if err := email.Press("Enter"); err != nil {
		return err
	}
	pwd := page.Locator("#password")
	fmt.Println("Filling in password field")
	if err := pwd.Click(); err != nil {
		return err
	} else if err := pwd.Fill(password); err != nil {
		return err
	} else if err := pwd.Press("Enter"); err != nil {
		return err
	}
	fmt.Println("Waiting for login to complete")
	if err := page.WaitForURL("https://account.venmo.com/"); err != nil {
		fmt.Println(page.URL())
		return err
	}

	return nil
}

func (v *Client) IsLoggedIn()(bool, error) {
	cookies, err := v.page.Context().Cookies()
	if err != nil {
		return false, err
	}
	for _, c := range cookies {
		if c.Name == "api_access_token" {
			return true, nil
		}
	}
	return false, nil
}
