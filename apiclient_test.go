package apiclient_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	. "github.com/happilymarrieddad/coinbase-v3-apiclient"
	"github.com/happilymarrieddad/coinbase-v3-apiclient/utils"

	cbadvclient "github.com/QuantFu-Inc/coinbase-adv/client"
	"github.com/QuantFu-Inc/coinbase-adv/model"
	coinbasegoclientv3 "github.com/happilymarrieddad/coinbase-go-client-v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("apiclient", func() {
	var (
		cont        ApiClient
		baseTicker  string
		quoteTicker string

		testBuy        bool
		testMarketBuy  bool
		testSell       bool
		testMarketSell bool

		fullTest bool
	)

	BeforeEach(func() {
		key, secret := os.Getenv("COINBASE_TEST_API_KEY"), os.Getenv("COINBASE_TEST_API_SECRET")

		// TODO: remove this requirement when coinbase adv adds support for all endpoints
		cbc, err := coinbasegoclientv3.NewClient(&http.Client{Timeout: time.Second * 30}, key, secret)
		Expect(err).To(BeNil())

		cont, err = NewApiClient(cbadvclient.NewClient(&cbadvclient.Credentials{ApiKey: key, ApiSKey: secret}), cbc, true)
		Expect(err).To(BeNil())

		Expect(cont).NotTo(BeNil())

		baseTicker = "YFI"
		quoteTicker = "BTC"
	})

	Context("GetCurrentWallentAmount", func() {
		It("should successfully verify current wallet amount", func() {
			baseAccount, quoteAccount, _, quoteAmt, err := cont.GetCurrentWallentAmount(ctx, baseTicker, quoteTicker)
			Expect(err).To(BeNil())
			// fmt.Println(baseAmt)
			// fmt.Println(quoteAmt)
			// Need the quote amount to be greater than 0 in order to cont the tests
			Expect(quoteAmt).To(BeNumerically(">", 0))
			Expect(baseAccount).NotTo(BeNil())
			Expect(quoteAccount).NotTo(BeNil())
		})
	})

	Context("GetProductMarketData", func() {
		It("should successfully get the current market product data", func() {
			high, low, price, _, err := cont.GetProductMarketData(ctx, baseTicker, quoteTicker, utils.Float64ToFloat64Ptr(0.1))
			Expect(err).To(BeNil())
			Expect(high).To(BeNumerically(">", 0))
			Expect(low).To(BeNumerically(">", 0))
			Expect(price).To(BeNumerically(">", 0))
		})
	})

	Context("CreateLimitMarketOrder", func() {
		Context("Buy", func() {
			BeforeEach(func() {
				if !testBuy {
					Skip("skipping buy tests")
				}
			})

			It("should successfully create a buy limit market order", func() {
				high, low, price, _, err := cont.GetProductMarketData(ctx, baseTicker, quoteTicker, utils.Float64ToFloat64Ptr(0.1))
				Expect(err).To(BeNil())
				fmt.Printf("High: %f Low: %f Price: %f\n", high, low, price)
				Expect(high).To(BeNumerically(">", 0))
				Expect(low).To(BeNumerically(">", 0))
				Expect(price).To(BeNumerically(">", 0))

				//expectedPerDiff := 0.5 // this is in whole floats (1 == 1%)

				By("using these numbers we are going to create a example buy")
				By("we want at least a 1% difference from the high and the low")
				priceDiffPercFromLow := (100 - ((low / price) * 100)) // this is a 5 for 5% at the moment
				//Expect(priceDiffPercFromLow).To(BeNumerically(">", expectedPerDiff))

				priceDiffPercFromHigh := (100 - ((price / high) * 100)) // this is a 1.9 for 1.9% at the moment
				//Expect(priceDiffPercFromHigh).To(BeNumerically(">", expectedPerDiff))

				fmt.Println(priceDiffPercFromLow, priceDiffPercFromHigh)
				By("coinbase fees we need at least a 1% difference in both high and low")

				// buyPrice := ((100 + expectedPerDiff) / 100) * price
				// sellPrice := ((100 - expectedPerDiff) / 100) * price

				buyPrice := price - ((price - low) / 2)
				sellPrice := ((high-price)/2 + price)

				fmt.Println("Buy Price: ", buyPrice, " Sell Price: ", sellPrice, " Current Price: ", price, " High: ", high, " Low: ", low)
				By("now we have the buy and sell price we need to get the amount we can buy/sell")

				// For now we only need the quote amount
				_, _, _, quoteAmount, err := cont.GetCurrentWallentAmount(ctx, baseTicker, quoteTicker)
				Expect(err).To(BeNil())
				amountWeCanBuy := (quoteAmount / buyPrice) * 0.995 // The 0.98 is to ensure there's enough quantity to make the buy
				fmt.Printf(
					"We can buy %f of %s at the %s price of %f with %f of %s\n",
					amountWeCanBuy, baseTicker, quoteTicker, buyPrice, quoteAmount, quoteTicker,
				)

				order, err := cont.CreateLimitMarketOrder(ctx, &CreateLimitMarketOrderParams{
					BaseTicker:  baseTicker,
					QuoteTicker: quoteTicker,
					Price:       buyPrice,
					Quantity:    amountWeCanBuy,
					Side:        BuySideType,
				})
				Expect(err).To(BeNil())
				Expect(order).NotTo(BeNil())
				fmt.Println("OrderID: ", order.GetOrderId())

				Expect(cont.CancelOrders(ctx, order.GetOrderId())).To(Succeed())
			})
		})

		Context("MarketBuy", func() {
			BeforeEach(func() {
				if !testMarketBuy {
					Skip("skipping market buy tests")
				}
			})

			It("should successfully create the order and then we'll cancel right away", func() {
				_, low, _, _, err := cont.GetProductMarketData(ctx, baseTicker, quoteTicker, utils.Float64ToFloat64Ptr(0.1))
				Expect(err).To(BeNil())

				_, _, _, quoteAmount, err := cont.GetCurrentWallentAmount(ctx, baseTicker, quoteTicker)
				Expect(err).To(BeNil())
				amountWeCanBuy := (quoteAmount / low) * 0.995

				orderID, err := cont.CreateOrderAndWaitForCompletion(ctx, &CreateLimitMarketOrderParams{
					BaseTicker:  baseTicker,
					QuoteTicker: quoteTicker,
					Price:       low,
					Quantity:    amountWeCanBuy,
					Side:        BuySideType,
				}, time.Now().Add(time.Second))
				Expect(err).NotTo(Succeed())
				Expect(err.Error()).To(Equal(fmt.Sprintf("market order '%s' has timed out", orderID)))

				Expect(cont.CancelOrders(ctx, orderID)).To(Succeed())
			})
		})

		Context("Sell", func() {
			BeforeEach(func() {
				if !testSell {
					Skip("skipping sell tests")
				}
			})

			It("should successfully create a buy limit market order", func() {
				high, low, price, _, err := cont.GetProductMarketData(ctx, baseTicker, quoteTicker, utils.Float64ToFloat64Ptr(0.1))
				Expect(err).To(BeNil())
				fmt.Printf("High: %f Low: %f Price: %f\n", high, low, price)
				Expect(high).To(BeNumerically(">", 0))
				Expect(low).To(BeNumerically(">", 0))
				Expect(price).To(BeNumerically(">", 0))

				//expectedPerDiff := 0.5 // this is in whole floats (1 == 1%)

				By("using these numbers we are going to create a example buy")
				By("we want at least a 1% difference from the high and the low")
				priceDiffPercFromLow := (100 - ((low / price) * 100)) // this is a 5 for 5% at the moment
				//Expect(priceDiffPercFromLow).To(BeNumerically(">", expectedPerDiff))

				priceDiffPercFromHigh := (100 - ((price / high) * 100)) // this is a 1.9 for 1.9% at the moment
				//Expect(priceDiffPercFromHigh).To(BeNumerically(">", expectedPerDiff))

				fmt.Println(priceDiffPercFromLow, priceDiffPercFromHigh)
				By("coinbase fees we need at least a 1% difference in both high and low")

				// buyPrice := ((100 + expectedPerDiff) / 100) * price
				// sellPrice := ((100 - expectedPerDiff) / 100) * price

				buyPrice := price - ((price - low) / 2)
				sellPrice := ((high-price)/2 + price)

				fmt.Println("Buy Price: ", buyPrice, " Sell Price: ", sellPrice, " Current Price: ", price, " High: ", high, " Low: ", low)
				By("now we have the buy and sell price we need to get the amount we can buy/sell")

				// For now we only need the quote amount
				_, _, baseAmount, _, err := cont.GetCurrentWallentAmount(ctx, baseTicker, quoteTicker)
				Expect(err).To(BeNil())
				fmt.Printf(
					"We can sell %f of %s at the %s price of %f\n",
					baseAmount, baseTicker, quoteTicker, sellPrice,
				)

				order, err := cont.CreateLimitMarketOrder(ctx, &CreateLimitMarketOrderParams{
					BaseTicker:  baseTicker,
					QuoteTicker: quoteTicker,
					Price:       sellPrice,
					Quantity:    baseAmount,
					Side:        SellSideType,
				})
				Expect(err).To(BeNil())
				Expect(order).NotTo(BeNil())
				fmt.Println("OrderID: ", order.GetOrderId())

				Expect(cont.CancelOrders(ctx, order.GetOrderId())).To(Succeed())
			})
		})

		Context("MarketSell", func() {
			BeforeEach(func() {
				if !testMarketSell {
					Skip("skipping market sell tests")
				}
			})

			It("should successfully create the order and then we'll cancel right away", func() {
				high, _, _, _, err := cont.GetProductMarketData(ctx, baseTicker, quoteTicker, utils.Float64ToFloat64Ptr(0.1))
				Expect(err).To(BeNil())

				_, _, baseAmount, _, err := cont.GetCurrentWallentAmount(ctx, baseTicker, quoteTicker)
				Expect(err).To(BeNil())

				orderID, err := cont.CreateOrderAndWaitForCompletion(ctx, &CreateLimitMarketOrderParams{
					BaseTicker:  baseTicker,
					QuoteTicker: quoteTicker,
					Price:       high,
					Quantity:    baseAmount,
					Side:        SellSideType,
				}, time.Now().Add(time.Second))
				Expect(err).NotTo(Succeed())
				Expect(err.Error()).To(Equal(fmt.Sprintf("market order '%s' has timed out", orderID)))

				Expect(cont.CancelOrders(ctx, orderID)).To(Succeed())
			})
		})

		Context("FullTest", func() {
			BeforeEach(func() {
				if !fullTest {
					Skip("skipping full test")
				}
			})

			It("should successfully buy into the coin", func() {
				_, _, price, _, err := cont.GetProductMarketData(ctx, baseTicker, quoteTicker, utils.Float64ToFloat64Ptr(0.1))
				Expect(err).To(BeNil())

				priceToBuy := price

				_, _, _, quoteAmount, err := cont.GetCurrentWallentAmount(ctx, baseTicker, quoteTicker)
				Expect(err).To(BeNil())
				amountWeCanBuy := (quoteAmount / priceToBuy) * 0.2

				orderID, err := cont.CreateOrderAndWaitForCompletion(ctx, &CreateLimitMarketOrderParams{
					BaseTicker:  baseTicker,
					QuoteTicker: quoteTicker,
					Price:       priceToBuy,
					Quantity:    amountWeCanBuy,
					Side:        BuySideType,
				}, time.Now().Add(time.Minute*5))
				if err != nil {
					log.Println("Order failed to purchase so cancelling the order")
					descr := "oh no... this should never happen... attempting to cancel order. look at your coinbase account IMMEDIATLY!!"
					Expect(cont.CancelOrders(ctx, orderID)).To(Succeed(), descr)
					// Forcing a fail to cancel the tests
					Expect(err).To(BeNil(), descr)
				}
				Expect(err).To(BeNil())

				By("we've successfully purchased some coins. Now lets sell them back")
				_, _, price, _, err = cont.GetProductMarketData(ctx, baseTicker, quoteTicker, utils.Float64ToFloat64Ptr(0.1))
				Expect(err).To(BeNil())

				priceToSell := price

				_, _, baseAmount, _, err := cont.GetCurrentWallentAmount(ctx, baseTicker, quoteTicker)
				Expect(err).To(BeNil())

				orderID, err = cont.CreateOrderAndWaitForCompletion(ctx, &CreateLimitMarketOrderParams{
					BaseTicker:  baseTicker,
					QuoteTicker: quoteTicker,
					Price:       priceToSell,
					Quantity:    baseAmount,
					Side:        SellSideType,
				}, time.Now().Add(time.Minute*5))
				if err != nil {
					log.Println("Order failed to sell so cancelling the order")
					descr := "oh no... this should never happen... attempting to cancel order. look at your coinbase account IMMEDIATLY!!"
					Expect(cont.CancelOrders(ctx, orderID)).To(Succeed(), descr)
					// Forcing a fail to cancel the tests
					Expect(err).To(BeNil(), descr)
				}
				Expect(err).To(BeNil())
			})
		})
	})

	Context("GetOrders", func() {
		It("should successfully get orders", func() {
			_, err := cont.GetOpenOrdersByProductIDAndSide(ctx, "OGN-BTC", model.BUY)
			Expect(err).To(BeNil())
		})
	})
})
