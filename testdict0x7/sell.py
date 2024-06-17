import ccxt
# importing datetime module for now()
import datetime
import asyncio
import json
import requests
import time
import sys


# KUCOINAPIKEY = sys.argv[1]
# KUCOINSECRET = sys.argv[2]  
# KUCOINPASSWORD = sys.argv[3]
KUCOINAPIKEY = "661bd98603e77600013bfd3d"
KUCOINSECRET = "7a0c92b9-69dc-40c0-8a01-30d73f660560"
#KUCOINPASSWORD = "47H6PJsb-5WRidM"
KUCOINPASSWORD = "*9Sd49G.!rt4RC$"
dollars_per_trade = 3
kucoin = ccxt.kucoin()

authentification = {
    "apiKey": f"{KUCOINAPIKEY}",
    "secret": f"{KUCOINSECRET}",
    "password": f"{KUCOINPASSWORD}",
}
kucoin = ccxt.kucoin(authentification)

def sell(coin, number_of_coins_to_sell):
    order_type = 'market'
    side = 'sell'
    current_price = (kucoin.fetch_ticker(coin)['ask'] + kucoin.fetch_ticker(coin)['bid'] )/ 2
    value = current_price * float(number_of_coins_to_sell)
    profit = value - 3
    kucoin.create_order(coin, order_type, side, number_of_coins_to_sell)
    print(f"Sold {number_of_coins_to_sell} {coin} for {current_price} each")
    print("Made a profit of ", profit)


coin_to_be_sold = "KARATE-USDT"
nuumber_of_coins_to_sell = "2730.2134"
sell(coin=coin_to_be_sold, number_of_coins_to_sell=nuumber_of_coins_to_sell)

