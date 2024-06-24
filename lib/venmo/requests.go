package venmo

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
)

type PendingRequests struct {
	Payments []PendingRequest
}
type PendingRequest struct {
	Type   string
	Status string
	Title  struct {
		TitleType string
		Payload   struct {
			Action  string
			SubType string
		}
		Receiver struct {
			Id          string
			DisplayName string
			Username    string
		}
	}
	Amount float64
	Note   struct {
		Content string
	}
}

func (v *Client) GetPendingRequests() ([]PendingRequest, error) {
	bytes, err := v.get("https://account.venmo.com/api/payments/outgoing?userId=2579504702685184569&action=charge&status=held,pending")
	if err != nil {
		return nil, err
	}
	requests := PendingRequests{}
	if err = json.Unmarshal(bytes, &requests); err != nil {
		return nil, err
	}
	return requests.Payments, nil
}

const batchSize = 10

// negative amount sends a payment, positive amount sends a request
func (v *Client) ProcessPayment(amt float64, note string, venmos []string) error {
	var buttonFn func() error
	if amt >= 0.005 {
		buttonFn = v.clickRequestButton
	} else if amt <= -0.005 {
		buttonFn = v.clickPayButton
	} else {
		return nil
	}

	if err := v.goToUrl("https://account.venmo.com/pay"); err != nil {
		return err
	}
	for _, venmo := range venmos {
		if err := cmp.Or(v.searchPaymentVenmo(venmo), v.clickSearchedPaymentVenmo(venmo)); err != nil {
			return err
		}
	}

	return cmp.Or(
		v.setPaymentAmount(amt),
		v.setPaymentNote(note),
		buttonFn(),
		v.waitForUrl("https://account.venmo.com/"),
	)
}

func outToFile(filepath string, content []byte) {
	fmt.Println("writing to file", filepath)
	file, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		panic(err)
	}
}
