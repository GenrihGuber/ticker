package main

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"strconv"
	"log"
	"sync"
	"io/ioutil"
)

type Bitpay struct {
	Rate float64
}

type Bitstamp struct {
	Last string
}

type Blockchain struct {
	USD struct {
		Last float64
	}
}

type Fixer struct {
	Rates struct {
		USD float64
	}
}

type FreeCC struct {
	EUR_USD struct {
		Val float64
	}
}

func bitpaySearch(c chan float64, url string) {
	data, err := fetch(url)
	if err != nil {
		log.Println(err)
	} else {
		var bitpay Bitpay
		err := json.Unmarshal(data, &bitpay)
		if err != nil {
			log.Println(err)
		} else {
			c <- bitpay.Rate
		}
	}
}

func bitstampSearch(c chan float64, url string) {
	data, err := fetch(url)
	if err != nil {
		log.Println(err)
	} else {
		var bitstamp Bitstamp
		err := json.Unmarshal(data, &bitstamp)
		if err != nil {
			log.Println(err)
		} else {
			rate, err := strconv.ParseFloat(bitstamp.Last, 64)
			if err != nil {
				log.Println(err)
			} else {
				c <- rate
			}
		}
	}
}

func blockchainSearch(c chan float64, url string) {
	data, err := fetch(url)
	if err != nil {
		log.Println(err)
	} else {
		var blockchain Blockchain
		err := json.Unmarshal(data, &blockchain)
		if err != nil {
			log.Println(err)
		} else {
			c <- blockchain.USD.Last
		}
	}
}

func fixerSearch(c chan float64 , url string) {
	data, err := fetch(url)
	if err != nil {
		log.Println(err)
	} else {
		var fixer Fixer
		err := json.Unmarshal(data, &fixer)
		if err != nil {
			log.Println(err)
		} else {
			c <- fixer.Rates.USD
		}
	}
}

func freeCCSearch(c chan float64, url string) {
	data, err := fetch(url)
	if err != nil {
		log.Println(err)
	} else {
		var freeCC FreeCC
		err := json.Unmarshal(data, &freeCC)
		if err != nil {
			log.Println(err)
		} else {
			c <- freeCC.EUR_USD.Val
		}
	}
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var respBody, respErr = ioutil.ReadAll(resp.Body)
	if respErr != nil {
		return nil, err
	}
	return respBody, nil
}

func runSearches(timeoutInMillis time.Duration) {
	ch_btc_usd := make(chan float64)
	ch_eur_usd := make(chan float64)

	go bitpaySearch(ch_btc_usd, "https://bitpay.com/api/rates/USD")
	go bitstampSearch(ch_btc_usd, "https://www.bitstamp.net/api/v2/ticker/btcusd")
	go blockchainSearch(ch_btc_usd, "https://blockchain.info/ticker")
	go bitstampSearch(ch_eur_usd, "https://www.bitstamp.net/api/v2/ticker/eurusd")
	go fixerSearch(ch_eur_usd, "http://api.fixer.io/latest?symbols=USD")
	go freeCCSearch(ch_eur_usd, "http://free.currencyconverterapi.com/api/v3/convert?q=EUR_USD&compact=y")

	timeout := time.After(timeoutInMillis * time.Millisecond)
	sum_btc_usd, sum_eur_usd := 0.0, 0.0
	n1, n2 := 0, 0

	loop:
	for i := 0; i < 6; i++ {
		select {
		case result := <-ch_btc_usd:
			sum_btc_usd += result
			n1++
		case result := <-ch_eur_usd:
			sum_eur_usd += result
			n2++
		case <-timeout:
			break loop
		}
	}

	btc_usd, eur_usd, btc_eur := 0.0, 0.0, 0.0
	if (n1 > 0) {
		btc_usd = sum_btc_usd / float64(n1)
	}
	if (n2 > 0) {
		eur_usd = sum_eur_usd / float64(n2)
	}
	if (n1 > 0 && n2 > 0) {
		btc_eur = btc_usd / eur_usd
	}
	fmt.Printf(" BTC/USD: %.2f EUR/USD: %.2f BTC/EUR: %.2f Active sources: BTC/USD (%v of %v) EUR/USD (%v of %v)\n",
		btc_usd, eur_usd, btc_eur, n1, 3, n2, 3)
}

func Ticker(tickInSeconds time.Duration, timeoutInMillis time.Duration, wg *sync.WaitGroup) {
	tick := time.Tick(tickInSeconds * time.Second)
	for {
		select {
		case <-tick:
			runSearches(timeoutInMillis)
		}
	}
	wg.Done()
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go Ticker(1, 500, &wg)
	wg.Add(1)
	go Ticker(5, 200, &wg)
	wg.Wait()
}
