package venmo

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

func init() {
	err := playwright.Install()
	if err != nil {
		panic(fmt.Sprintf("error installing playwright: %s", err))
	}
}