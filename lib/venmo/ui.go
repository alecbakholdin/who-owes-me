package venmo

import (
	"cmp"
	"fmt"
	"math"
	"strings"
)

func (v *Client) goToUrl(url string) error {
	fmt.Println("Going to", url)
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		_, err = page.Goto(url)
		return err
	}
}

func (v *Client) waitForUrl(url string) error {
	fmt.Println("Waiting for", url)
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		return page.WaitForURL(url)
	}
}

func (v *Client) formatVenmo(venmo string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(venmo), "@"))
}

func (v *Client) searchPaymentVenmo(venmo string) error {
	formattedVenmo := v.formatVenmo(venmo)
	fmt.Println("searching venmo", formattedVenmo)
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		searchInput := page.Locator("#search-input")
		return cmp.Or(
			searchInput.Click(),
			searchInput.Fill(formattedVenmo),
		)
	}
}

func (v *Client) clickSearchedPaymentVenmo(venmo string) error {
	formattedVenmo := v.formatVenmo(venmo)
	fmt.Println("Clicking venmo", formattedVenmo)
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		venmoLiItem := page.Locator(fmt.Sprintf("li img[alt='%s'i]", formattedVenmo))
		return cmp.Or(
			venmoLiItem.WaitFor(),
			venmoLiItem.Click(),
		)
	}
}

func (v *Client) setPaymentAmount(amt float64) error {
	fmt.Println("setting amount", math.Abs(amt))
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		amountInput := page.Locator("form input[placeholder='0']")
		return cmp.Or(
			amountInput.Click(),
			amountInput.Fill(fmt.Sprintf("%.02f", math.Abs(amt))),
		)
	}
}

func (v *Client) setPaymentNote(note string) error {
	fmt.Println("setting note")
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		noteInput := page.Locator("#payment-note")
		return cmp.Or(
			noteInput.Click(),
			noteInput.Fill(note),
		)
	}
}

func (v *Client) clickRequestButton() error {
	fmt.Println("clicking request")
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		return cmp.Or(
			page.Locator(".pay_payRequestButton__6ao19>button:last-child").Click(),
			page.Locator("form>button[type='submit']:not([disabled])").Click(),
		)
	}
}

func (v *Client) clickPayButton() error {
	fmt.Println("clicking pay")
	if page, err := v.initPage(); err != nil {
		return err
	} else {
		return cmp.Or(
			page.Locator(".pay_payRequestButton__6ao19>button:first-child").Click(),
			page.Locator("form div[class^='fundingInstrument'] button[type='button']:not([disabled]):first-child").Click(),
		)
	}
}
