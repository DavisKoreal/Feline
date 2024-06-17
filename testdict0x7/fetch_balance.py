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
kucoin.fetch_balance()

