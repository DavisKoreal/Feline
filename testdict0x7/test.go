package main

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"encoding/json"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	//"net/http"
	"os/exec"
	"sort"
	"time"

	"github.com/Kucoin/kucoin-go-sdk"

	"log"

	"os"
	// "net/http"
	// "strings"
	//"github.com/rs/zerolog/log"
	//"github.com/go-gota/gota/dataframe"
	//"github.com/go-gota/gota/series"
)

const (
	ChannelTicker   string = "ticker"
	TypeSubscribe   string = "subscribe"
	TypeUnsubscribe string = "unsubscribe"
)

type Message struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Topic    string `json:"topic"`
	Response bool   `json:"responseonse"`
}

type Trade struct {
	Type   string    `json:"type"`
	Time   time.Time `json:"time"`
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
}

//{"topic":"/market/ticker:all","type":"message","data":{"bestAsk":"0.17257","bestAskSize":"2311.8817","bestBid":"0.17151",
//"bestBidSize":"2150.1183","price":"0.17157","sequence":"533524567","size":"43.4673","time":1714317416582},"subject":"UOS-USDT"}

type data struct {
	BestAsk     string `json:"bestAsk"`
	BestAskSize string `json:"bestAskSize"`
	BestBid     string `json:"bestBid"`
	BestBidSize string `json:"bestBidSize"`
	Price       string `json:"price"`
	Sequence    string `json:"sequence"`
	Size        string `json:"size"`
	Time        int    `json:"time"`
}

type response_struct struct {
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Data    data   `json:"data"`
	Subject string `json:"subject"`
}

type state_variables struct {
	Subject      string
	Price        float64
	Volume       float64
	Pricechange  float64
	Volumechange float64
	Time         int
}

type bought_coins struct {
	Subject       string
	number_bought float32
}

type investment_track struct {
	Subject     string
	Percentgain float64
	start_price float64
}

func this_coin_is_usdt(subject string) bool {
	return subject[len(subject)-4:] == "USDT"
}

// type row struct {
// 	Subject string
// 	Price float64
// 	Volume float64
// 	Pricechange float64
// 	Volumechange float64

// }

// KUCOINAPIKEY = "661bd98603e77600013bfd3d"
// KUCOINSECRET = "7a0c92b9-69dc-40c0-8a01-30d73f660560"
// #KUCOINPASSWORD = "47H6PJsb-5WRidM"
// KUCOINPASSWORD = "*9Sd49G.!rt4RC$"
var KUCOINAPIKEY = "661bd98603e77600013bfd3d"
var KUCOINSECRET = "7a0c92b9-69dc-40c0-8a01-30d73f660560"
var KUCOINPASSWORD = "*9Sd49G.!rt4RC$"
var tradingisallowed = false

// var timetoclosealltrades = false
var printunreasonables = false

// var tradingdollars string = "1"
var tradedollarsfloat float64 = 3
var target_percentage float64 = 30
var connection_websocket = false

var Adress string = "wss://ws-api.kucoin.com/?token=2neAiuYvAU61ZDXANAGAsiL4-iAExhsBXZxftpOeh_55i3Ysy2q2LEsEWU64mdzUOPusi34M_wGoSf7iNyEWJ1pHBCi_DYtWzyUy2oXl06XtvPiaVhuJ29iYB9J6i9GjsxUuhPw3BlrzazF6ghq4L3zxm0ToCSIdAD3cpGR2_eg=.C0a843NFzN_t-i0jC5q8Dw==&[connectId=Dave2024]"

// var number_of_top_to_buy int = 5
var tradinghour int = 16
var tradingminute int = 59
var tradingsecond int = 00
var filter_price_change float64 = 10
var negativepercentage float64 = -20
var buytrials = 0
var dollarsused float64 = 0.0
var maxnumberoftrades int = 5

// maps used to store data
var data_map = make(map[string]state_variables)
var purchase_map = make(map[string]bought_coins)
var tradestracking = make(map[string]investment_track)

// initiating variable s to be used in the runtime
var s *kucoin.ApiService = kucoin.NewApiService(
	kucoin.ApiKeyOption(KUCOINAPIKEY),
	kucoin.ApiSecretOption(KUCOINSECRET),
	kucoin.ApiPassPhraseOption(KUCOINPASSWORD),
)

//functions used in the runtime

func serverTime(s *kucoin.ApiService) {
	rsp, err := s.ServerTime()
	if err != nil {
		log.Printf("Error: %s", err.Error())
		// Handle error
		return
	}

	var ts int64
	if err := rsp.ReadData(&ts); err != nil {
		// Handle error
		return
	}
	log.Printf("The server time: %d", ts)
}

func round_to_one_decimal(number float64) float64 {
	return float64(int(number*10)) / 10
}

func coin_has_been_purchased(coin_to_buy string) bool {
	coin, coin_is_in_map := purchase_map[coin_to_buy]
	fmt.Println("The message is: ", coin)
	return coin_is_in_map
}

func buy(coin_to_buy string) {
	file, errs := os.OpenFile("buy.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errs != nil {
		fmt.Println("Failed to create file:", errs)
		return
	}
	defer file.Close()
	size_of_trade := round_to_one_decimal(tradedollarsfloat / data_map[coin_to_buy].Price)
	floatString := strconv.FormatFloat(float64(size_of_trade), 'f', -1, 64)
	p := &kucoin.CreateOrderModel{
		ClientOid: kucoin.IntToString(time.Now().UnixNano()),
		Side:      "buy",
		Symbol:    coin_to_buy,
		Type:      "market",
		Size:      floatString,
	}
	// Open the file for writing
	_, errs = file.WriteString("Buying " + floatString + " of " + coin_to_buy + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
	currenttime := time.Now()

	rsp, err := s.CreateOrder(p)
	if err != nil {
		fmt.Println(err)
		fmt.Println("The coin ", coin_to_buy, " has not been bought")
		return
	}
	for {
		if len(rsp.Message) > 1 {
			rsp, err = s.CreateOrder(p)
			if err != nil {
				fmt.Println(err)
				fmt.Println("The coin ", coin_to_buy, " has not been bought")
				return
			}
		}
		if len(rsp.Message) < 1 {
			break
		}
	}

	_, errs = file.WriteString(rsp.Message + "\n" + currenttime.String() + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
	//fmt.Println("response: ", rsp)
	dollarsused += 1
	fmt.Println(coin_to_buy, " has been bought")
}

func sell(coin_to_sell string, number_to_sell string) {

	file, errs := os.OpenFile("sell.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errs != nil {
		fmt.Println("Failed to create file:", errs)
		return
	}
	defer file.Close()
	p := &kucoin.CreateOrderModel{
		ClientOid: kucoin.IntToString(time.Now().UnixNano()),
		Side:      "sell",
		Symbol:    coin_to_sell,
		Type:      "market",
		Size:      number_to_sell,
	}
	_, errs = file.WriteString("Selling " + number_to_sell + " of " + coin_to_sell + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
	rsp, err := s.CreateOrder(p)
	if err != nil {
		fmt.Println(err)
		fmt.Println("The coin ", coin_to_sell, " has not been sold")
		return
	}
	_, errs = file.WriteString(rsp.Message + "\n")
	if errs != nil {
		fmt.Println("Failed to write to file:", errs) //print the failed message
		return
	}
	fmt.Println("response: ", rsp)
}

func track_running_trades(coin_to_track string, currentprice float64) {
	coin, coin_is_in_map := purchase_map[coin_to_track]
	//fmt.Println("The message is: ", coin)
	if coin_is_in_map {
		percentgain := (currentprice - tradestracking[coin_to_track].start_price) / tradestracking[coin_to_track].start_price * 100
		if percentgain > target_percentage {
			closealltrades()
			fmt.Println("All coins that have been sold have been closed because of the last stable change")
			return
		}
		if percentgain < negativepercentage {
			sell(coin_to_track, strconv.FormatFloat(float64(purchase_map[coin_to_track].number_bought), 'f', 6, 32))

		}
		if len(tradestracking) >= maxnumberoftrades {
			closealltrades()
			fmt.Println("All coins that have been sold have been closed because of the last stable change")
			return
		}
		tradestracking[coin_to_track] = investment_track{coin_to_track, percentgain, tradestracking[coin_to_track].start_price}
	}

	if !coin_is_in_map {
		fmt.Println("The coin ", coin, " is not in the purchase_map")
	}
}

func closealltrades() {
	for key := range purchase_map {
		sell(purchase_map[key].Subject, strconv.FormatFloat(float64(purchase_map[key].number_bought), 'f', 6, 32))
	}
	tradingisallowed = false
}

func printvariables() {
	fmt.Println("The number of records seen is: \t", records_seen)
	fmt.Println("Buy trials is: \t", buytrials)
	fmt.Println("Tradingallowed:\t ", tradingisallowed)
	fmt.Println("Connection is: ", connection_websocket)
	fmt.Println("Dollars used is: ", dollarsused)
	fmt.Println("Filter price change is: ", filter_price_change)
	fmt.Println("Target percentage is: ", target_percentage)
	fmt.Println("\nPurchase map is: ", purchase_map)
	fmt.Println("\n............")
}

func clear_terminal() {
	cmd := exec.Command("clear")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
	fmt.Println("Output: ", string(out))
}

func update_trading_is_allowed() {
	if !tradingisallowed {
		dt := time.Now()
		if (dt.Hour() == tradinghour && dt.Minute() >= tradingminute && dt.Second() >= tradingsecond) || (dt.Hour() > tradinghour) {
			tradingisallowed = true
		}

	}
}

// func get_the_token() string {
// 	// Post data to url
// 	var url string = "https://api.kucoin.com/api/v1/bullet-public"
// 	resp, err := http.Post(url, "application/json", nil)
// 	if err != nil {
// 		fmt.Println("Error posting to url: ", err)
// 	}
// 	res := (strings.Split((strings.Split((resp.Header.Values("Set-Cookie")[2]), ";")[0]), "ken="))[1]
// 	return res
// }

var records_seen int = 0

//var nrow row =
//var df dataframe.DataFrame= dataframe.LoadStructs([]row{ {"firstrand", 0.987, 2.3, 0.0, 0.0}})

func main() {
	serverTime(s)
	c, _, err := websocket.Dial(context.Background(), Adress, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("I am connected")
	connection_websocket = true

	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	_, message, err := c.Read(context.Background())
	fmt.Println("The type of message is:	", reflect.TypeOf(message))

	if err != nil {
		fmt.Println(err)
		// Adress = "wss://ws-api.kucoin.com/?token=" + get_the_token() + "&[connectId=Dave2024]"
		// main()
	}
	//fmt.Println("The message is: ", message)
	fmt.Println("Message: ", message)
	//fmt.Println("The message is " + string(message))

	//subscribing to a channel in the websocket by defining a message first annd connecting to channel
	sub := Message{
		Id:       "Dave2024",
		Type:     "subscribe",
		Topic:    "/market/ticker:all",
		Response: true,
	}
	//fmt.Println("sending the message to the channel")
	//fmt.Println("The message is: ", sub)
	//fmt.Println("The type of message is:	", reflect.TypeOf(sub))
	err = wsjson.Write(context.Background(), c, sub)
	if err != nil {
		fmt.Println(err)
		return
	}

	//reading the responseonse from the channel
	for i := 0; i < 10000000; i++ {

		//This cell clears the output of the terminal
		clear_terminal()

		//closes all trades and returns if it is time to close trades
		// if timetoclosealltrades {
		// 	closealltrades()
		// }

		//this updates the variable tradingisallowed
		update_trading_is_allowed()

		//his function prints the reasonable variables
		printvariables()

		//This cell gets the message from the channel and confirms that the message is received
		_, message, err := c.Read(context.Background())
		if err != nil {
			fmt.Println(err)
			main()
		}
		stringmessage := string(message)
		records_seen += 1

		//This cell reads the json message and prints the subject of the message
		var response response_struct
		err = json.Unmarshal([]byte(stringmessage), &response)
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("The message is: ", stringmessage)
			fmt.Println("The routine for reading a json has been triggered ")
			main()
			return
		}

		//This cell checks if the subject is USDT
		//If it is not a USDT coin, the loop continues
		if !this_coin_is_usdt(response.Subject) {
			//fmt.Println("The subject ", response.Subject, " is not USDT")
			continue
		}

		//This cell creates a state variable for the coin
		statevol := 0.0
		stateprice, _ := strconv.ParseFloat(response.Data.Price, 64)
		statevolchange := 0.0
		statepricechange := 0.0
		statetime := time.Now().Nanosecond()
		state := state_variables{response.Subject, stateprice, statevol, statepricechange, statevolchange, statetime}
		//fmt.Println("state: ", state)

		//This cell checks if the coin is in the data_map
		coin, coin_is_in_map := data_map[response.Subject]
		// fmt.Println("the time is ", response.Data.Time)
		// fmt.Println("THe type of time is ", reflect.TypeOf(response.Data.Time))

		//printunreasonables if for printing unused variables so hat I do not
		//have a hard time during compilinf of the go program
		if printunreasonables {
			fmt.Println("The message is: ", coin)
		}

		//This cell updates the state of the coin in the data_map because the coin has been proven to have been
		//seen before
		if coin_is_in_map {

			//This cell checks if trading is allowed and if it is, it tracks the running trades
			//THe cell tracks the running trades of the currently accesed coin
			if tradingisallowed {
				track_running_trades(response.Subject, state.Price)
			}
			//fmt.Println("data map time : ", data_map[response.Subject].Time)

			//fmt.Println(state.Time," - ", data_map[response.Subject].Time, " = ", state.Time - data_map[response.Subject].Time)
			//fmt.Println(response.Subject, " \tis in the data_map")
			//fmt.Println("The coin ", response.Subject, " has the following data from coin variable: ", coin)
			//fmt.Println("The coin ", response.Subject, " has the following data from datamap : ", data_map[response.Subject])
			//fmt.Println("THe current state of ", response.Subject, " is ", state)
			state.Pricechange = (((state.Price - data_map[response.Subject].Price) / data_map[response.Subject].Price * 100) / float64((state.Time - data_map[response.Subject].Time))) * 10000000000

			state.Volumechange = (state.Volume - data_map[response.Subject].Volume) / data_map[response.Subject].Volume * 100

			data_map[response.Subject] = state
			//data_map[response.Subject].Time = state.Time

			keys := make([]string, 0, len(data_map))
			//fmt.Println("Getting keys from the data_map")
			for key := range data_map {
				keys = append(keys, key)
			}
			//fmt.Println("Now slicing and sorting ")
			sort.SliceStable(keys, func(i, j int) bool {
				return data_map[keys[i]].Pricechange > data_map[keys[j]].Pricechange
			})
			//fmt.Println("THe type of keys is ", reflect.TypeOf(keys))
			//fmt.Println("NOw iterating based on odrdered list ")
			fmt.Println("There are ", len(keys), " keys in the data_map")
			for key := 0; key < len(keys) && key < 10; key++ {
				fmt.Println(data_map[keys[key]].Subject, "\t\t\t", data_map[keys[key]].Price, "\t\t\t ", data_map[keys[key]].Pricechange)
				//, "\t ", data_map[keys[key]].Volume, "\t ", data_map[keys[key]].Volumechange

				//if it is tradingisallowed
				if tradingisallowed && (data_map[keys[key]].Pricechange > filter_price_change) {
					if !coin_has_been_purchased(data_map[keys[key]].Subject) {
						go buy(data_map[keys[key]].Subject)
						ammount := round_to_one_decimal(tradedollarsfloat / data_map[keys[key]].Price)
						purchase_map[data_map[keys[key]].Subject] = bought_coins{data_map[keys[key]].Subject, float32(ammount)}
						buytrials += 1
						tradestracking[data_map[keys[key]].Subject] = investment_track{data_map[keys[key]].Subject, 0.0, data_map[keys[key]].Price}
					}
					//fmt.Println("The coin ", data_map[keys[key]].Subject, " has been bought at ", data_map[keys[key]].Price)

				}
			}
			fmt.Println("........")
			fmt.Println("........")
			for key := len(keys) - 11; key < len(keys) && key > 0; key++ {
				fmt.Println(data_map[keys[key]].Subject, "\t\t\t", data_map[keys[key]].Price, "\t\t\t ", data_map[keys[key]].Pricechange)
				//, "\t ", data_map[keys[key]].Volume, "\t ", data_map[keys[key]].Volumechange)
			}

			continue
		}

		//This cell updates the data_map with the new coin and its state
		data_map[response.Subject] = state

		//This cell ranks the ranking list of keys
		keys := make([]string, 0, len(data_map))
		//fmt.Println("Getting keys from the data_map")
		for key := range data_map {
			keys = append(keys, key)
		}
		//fmt.Println("Now slicing and sorting ")
		sort.SliceStable(keys, func(i, j int) bool {
			return data_map[keys[i]].Pricechange > data_map[keys[j]].Pricechange
		})
		// fmt.Println("Ordered keys are: ", keys)
		// fmt.Println("THe type of keys is ", reflect.TypeOf(keys))
		// fmt.Println("Ordered data_map is: ", data_map)

		//This cell prints the top 10 coins in the data_map
		//fmt.Println("NOw iterating based on odrdered list ")
		fmt.Println("There are ", len(keys), " keys in the data_map")
		for key := 0; key < len(keys) && key < 10; key++ {
			fmt.Println(data_map[keys[key]].Subject, "\t\t\t", data_map[keys[key]].Price, "\t\t\t ", data_map[keys[key]].Pricechange)
			//, "\t ", data_map[keys[key]].Volume, "\t ", data_map[keys[key]].Volumechange
		}
		fmt.Println("........")
		fmt.Println("........")
		for key := len(keys) - 11; key < len(keys) && key > 0; key++ {
			fmt.Println(data_map[keys[key]].Subject, "\t\t\t", data_map[keys[key]].Price, "\t\t\t ", data_map[keys[key]].Pricechange)
			//, "\t ", data_map[keys[key]].Volume, "\t ", data_map[keys[key]].Volumechange)
		}
	}

}
