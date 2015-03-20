// Forex data API

// http://finance.yahoo.com/webservice/v1/symbols/CNY=X/quote?format=json

package forex

import (
	"encoding/json"
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"net/http"
	"time"
)

// Forex data API URL
const (
	DATAURL = "http://finance.yahoo.com/webservice/v1/symbols/"
)

// Quote forex info
type Quote struct {
	Price  float64
	Symbol string
	Error  error
}

// CommunicateFX sends the latest FX quote to the supplied channel
func CommunicateFX(symbol string, fxChan chan<- Quote, doneChan <-chan bool) error {
	// Check connection is ok to start
	quote := getQuote(symbol)
	if quote.Error != nil {
		return quote.Error
	}
	if quote.Price == 0 {
		return fmt.Errorf("Price is zero")
	}

	// Run read loop in new goroutine
	go runLoop(symbol, fxChan, doneChan)
	return nil
}

// HTTP read loop
func runLoop(symbol string, fxChan chan<- Quote, doneChan <-chan bool) {
	ticker := time.NewTicker(15 * time.Second)

	for {
		select {
		case <-doneChan:
			close(fxChan)
			return
		case <-ticker.C:
			fxChan <- getQuote(symbol)
		}
	}
}

// Returns quote for requested instrument
func getQuote(symbol string) Quote {
	tmp := struct {
		List struct {
			Resources []struct {
				Resource struct {
					Fields struct {
						Price float64 `json:"price,string"`
					} `json:"fields"`
				} `json:"resource"`
			} `json:"resources"`
		} `json:"list"`
	}{}

	url := fmt.Sprintf("%s=x/quote?format=json", symbol)

	data, err := get(url)
	if err != nil {
		return Quote{Error: fmt.Errorf("Forex error %s", err)}
	}

	if err = json.Unmarshal(data, &tmp); err != nil {
		return Quote{Error: fmt.Errorf("Forex error %s", err)}
	}

	return Quote{
		Price:  tmp.List.Resources[0].Resource.Fields.Price,
		Symbol: symbol,
		Error:  nil,
	}
}

// unauthenticated GET
func get(url string) ([]byte, error) {
	resp, err := http.Get(DATAURL + url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
