# What it is

Application created for code review

Unification of data retrieval from various crypto exchanges.

## Build:

1. clone repo: `git clone git@github.com:vs0x309/review.git`
2. cd review
3. go build .

## Run commands:

To get all commands, run the application with the -help parameter.

    -addr string
        server addres (default ":8080")
    -logFile string
        Path to log file

## Example:

1. `curl http://127.0.0.1:8080/exchanges`
2. `curl http://127.0.0.1:8080/bybit/pairs`
3. `curl http://127.0.0.1:8080/bybit/orderbook/BTCUSDT`
4. `curl http://127.0.0.1:8080/gateio/pairs`
5. `curl http://127.0.0.1:8080/gateio/orderbook/BTC_USDT`