package apiclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/happilymarrieddad/coinbase-v3-apiclient/utils"

	cbadvclient "github.com/QuantFu-Inc/coinbase-adv/client"
	"github.com/QuantFu-Inc/coinbase-adv/model"
	cbadvmodel "github.com/QuantFu-Inc/coinbase-adv/model"
	"github.com/google/uuid"
	coinbasegoclientv3 "github.com/happilymarrieddad/coinbase-go-client-v3"
)

type sideType string

const (
	BuySideType  sideType = "BUY"
	SellSideType sideType = "SELL"
)

//go:generate mockgen -destination=./mocks/CoinbaseClient.go -package=mocks github.com/QuantFu-Inc/coinbase-adv/client CoinbaseClient

//go:generate mockgen -destination=./mocks/ApiClient.go -package=mocks github.com/happilymarrieddad/coinbase-v3-apiclient ApiClient
type ApiClient interface {
	CreateOrderAndWaitForCompletion(ctx context.Context, params *CreateLimitMarketOrderParams, timeout time.Time) (orderID string, err error)

	// Helpers
	GetCurrentWallentAmount(
		ctx context.Context, baseTicker, quoteTicker string,
	) (baseAccount, quoteAccount *cbadvmodel.Account, baseAmount, quoteAmount float64, err error)
	GetProduct(ctx context.Context, baseTicker, quoteTicker string) (product *cbadvmodel.GetProductResponse, err error)
	GetProductMarketData(ctx context.Context, baseTicker, quoteTicker string, pricePercentageChange24h *float64) (highLast24Hr, lowLast24Hr, currentPrice, currentPriceChangePercentage float64, err error)
	CreateLimitMarketOrder(ctx context.Context, params *CreateLimitMarketOrderParams) (order *cbadvmodel.Order, err error)
	VerifyMarketOrderCompletion(ctx context.Context, orderID string, timeout time.Time) error
	GetOrder(ctx context.Context, orderID string) (*cbadvmodel.Order, error)
	GetOpenOrdersByProductIDAndSide(ctx context.Context, productID string, side cbadvmodel.OrderSide) ([]cbadvmodel.Order, error)
	GetOrderFills(ctx context.Context, orderID, productID string) ([]cbadvmodel.OrderFill, error)
	CancelOrders(ctx context.Context, orderIds ...string) (err error)
	CancelExistingOrders(ctx context.Context, id string, productID string, orderType model.OrderType) (err error)
}

// NewApiClient backup will not be needed once the main client supports MarketTrades and ListProducts
func NewApiClient(client cbadvclient.CoinbaseClient, backup coinbasegoclientv3.Client, debug bool) (ApiClient, error) {
	// forcing debug for now
	c := apiclient{client: client, backup: backup, mutex: &sync.RWMutex{}, debug: debug}

	return &c, nil
}

type apiclient struct {
	client               cbadvclient.CoinbaseClient
	backup               coinbasegoclientv3.Client
	accountUUIDsByTicker map[string]string
	mutex                *sync.RWMutex
	debug                bool
	// helper parameters
	hasMentionedOrderWaiting bool
}

func (c *apiclient) CreateOrderAndWaitForCompletion(ctx context.Context, params *CreateLimitMarketOrderParams, timeout time.Time) (orderID string, err error) {
	// 1 - create market order to buy the base ticker using quote ticker
	order, err := c.CreateLimitMarketOrder(ctx, params)
	if err != nil {
		return "", err
	}

	// 2 - either wait until market order completes or if timeout/error return
	if err = c.VerifyMarketOrderCompletion(ctx, order.GetOrderId(), timeout); err != nil {
		return order.GetOrderId(), err
	}

	return order.GetOrderId(), nil
}

func (c *apiclient) GetCurrentWallentAmount(
	ctx context.Context, baseTicker, quoteTicker string,
) (baseAccount, quoteAccount *cbadvmodel.Account, baseAmount, quoteAmount float64, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.accountUUIDsByTicker == nil {
		c.log("account data not available so fetching from remote service")
		c.accountUUIDsByTicker = make(map[string]string)

		accs, err := c.client.ListAccounts(&cbadvclient.ListAccountsParams{
			Limit: utils.Int32ToPtr(250), // TODO: this won't work when coinbase gets more than 250 coins...
		})
		if err != nil {
			c.log("client.ListAccounts err: %s", err.Error())
			return nil, nil, 0, 0, err
		}

		for _, acc := range accs.Accounts {
			c.accountUUIDsByTicker[utils.StringPtrToString(acc.Currency)] = utils.StringPtrToString(acc.Uuid)
		}

		c.log("finished fetching remote account data")
	}

	baseAccUUID, exists := c.accountUUIDsByTicker[baseTicker]
	if !exists {
		return nil, nil, 0, 0, fmt.Errorf("baseTicker '%s' account not available", baseTicker)
	}
	currentBaseAcc, err := c.client.GetAccount(baseAccUUID)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	quoteAccUUID, exists := c.accountUUIDsByTicker[quoteTicker]
	if !exists {
		return nil, nil, 0, 0, fmt.Errorf("quoteTicker '%s' account not available", quoteTicker)
	}
	currentQuoteAcc, err := c.client.GetAccount(quoteAccUUID)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return currentBaseAcc, currentQuoteAcc,
		utils.Float64PtrToFloat64(currentBaseAcc.AvailableBalance.Value),
		utils.Float64PtrToFloat64(currentQuoteAcc.AvailableBalance.Value), nil
}

func (c *apiclient) GetProduct(ctx context.Context, baseTicker, quoteTicker string) (product *cbadvmodel.GetProductResponse, err error) {
	return c.client.GetProduct(fmt.Sprintf("%s-%s", baseTicker, quoteTicker))
}

func (c *apiclient) GetProductMarketData(
	ctx context.Context, baseTicker, quoteTicker string, pricePercentageChange24h *float64,
) (highLast24Hr, lowLast24Hr, currentPrice, currentPriceChangePercentage float64, err error) {
	productID := fmt.Sprintf("%s-%s", baseTicker, quoteTicker)

	resPtr, err := c.client.GetProduct(productID)
	if err != nil {
		return 0, 0, 0, 0, err
	} else if resPtr == nil {
		// This should never happen
		return 0, 0, 0, 0, errors.New("response from client is nil")
	} else if pricePercentageChange24h != nil &&
		math.Abs(utils.Float64PtrToFloat64(resPtr.PricePercentageChange24h)) <= utils.Float64PtrToFloat64(pricePercentageChange24h) {
		// Not going to bother trying when the percentage is less than PricePercentageChange24h
		return 0, 0, 0, 0, fmt.Errorf("perc 24hr change less than %f%%", utils.Float64PtrToFloat64(pricePercentageChange24h))
	}

	trades, err := c.backup.GetMarketTrades(ctx, utils.StringPtrToString(resPtr.ProductId), 1000)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	for idx, trade := range trades {
		// This should never err
		price, strConvErr := strconv.ParseFloat(trade.Price, 64)

		if idx == 0 {
			if strConvErr != nil {
				return 0, 0, 0, 0, strConvErr
			}
			highLast24Hr, lowLast24Hr = price, price
		}

		if price > highLast24Hr {
			highLast24Hr = price
		}

		if price < lowLast24Hr {
			lowLast24Hr = price
		}
	}

	return highLast24Hr, lowLast24Hr, utils.Float64PtrToFloat64(resPtr.Price), utils.Float64PtrToFloat64(resPtr.PricePercentageChange24h), nil
}

type CreateLimitMarketOrderParams struct {
	ID          string
	BaseTicker  string   `validate:"required"`
	QuoteTicker string   `validate:"required"`
	Price       float64  `validate:"required"`
	Quantity    float64  `validate:"required"`
	Side        sideType `validate:"required"`
	// 1% will be 1.0
	PricePercentageChange24h *float64
	NumOfTries               int
}

// CreateLimitMarketOrder
//
//	side - BUY or SELL
func (c *apiclient) CreateLimitMarketOrder(ctx context.Context, params *CreateLimitMarketOrderParams) (*cbadvmodel.Order, error) {
	if err := utils.Validate(params); err != nil {
		return nil, err
	}

	if params.ID == "" {
		params.ID = uuid.New().String()
	}

	params.ID += fmt.Sprintf("-%d", time.Now().UnixMilli())

	coid := fmt.Sprintf("create-market-order-%s-%s-%s-%s", params.Side, params.BaseTicker, params.QuoteTicker, params.ID)
	productID := fmt.Sprintf("%s-%s", params.BaseTicker, params.QuoteTicker)

	req, err := c.client.CreateOrder(&cbadvmodel.CreateOrderRequest{
		ClientOrderId: utils.StringToPtr(coid),
		ProductId:     utils.StringToPtr(productID),
		Side:          utils.StringToPtr(string(params.Side)),
		OrderConfiguration: &cbadvmodel.CreateOrderRequestOrderConfiguration{
			LimitLimitGtc: &cbadvmodel.CreateOrderRequestOrderConfigurationLimitLimitGtc{
				BaseSize:   utils.StringToPtr(fmt.Sprintf("%f", params.Quantity)),
				LimitPrice: utils.StringToPtr(fmt.Sprintf("%f", params.Price)),
			},
		},
	})
	if err != nil {
		return nil, err
	} else if req == nil {
		return nil, errors.New("create order request is nil")
	} else if !*req.Success {
		if params.NumOfTries > 5 {
			c.log("Num of retries has exceeded the allowed amount so just returning the error")
			return nil, errors.New(req.ErrorResponse.GetMessage())
		}

		// This is the most dodgy place in the app...
		// Need to actually compare errors and handle properly
		if req.ErrorResponse.GetError() == "INVALID_PRICE_PRECISION" {
			newPrice := utils.TrimFloatToRight(params.Price, 1)
			c.log("invalid price precision. current price: %f, new price: %f", params.Price, newPrice)
			params.Price = newPrice
			time.Sleep(time.Millisecond * 50) // just delay slightly
			params.NumOfTries++
			return c.CreateLimitMarketOrder(ctx, params)
		}

		if req.ErrorResponse.GetError() == "INSUFFICIENT_FUND" {
			newQuantity := utils.TrimFloatToRight(params.Quantity, 1)
			c.log("invalid quantity. current quantity: %f, new quantity: %f", params.Quantity, newQuantity)
			params.Quantity = newQuantity
			time.Sleep(time.Millisecond * 50) // just delay slightly
			params.NumOfTries++
			return c.CreateLimitMarketOrder(ctx, params)
		}

		if req.ErrorResponse.GetError() == "INVALID_SIZE_PRECISION" {
			if req.ErrorResponse.GetMessage() == "Too many decimals in order amount" {
				newQuantity := utils.TrimFloatToRight(params.Quantity, 1)
				c.log("invalid quantity. current quantity: %f, new quantity: %f", params.Quantity, newQuantity)
				params.Quantity = newQuantity
				time.Sleep(time.Millisecond * 50) // just delay slightly
				params.NumOfTries++
				return c.CreateLimitMarketOrder(ctx, params)
			}
		}

		// This is just so I can add error handling as I go
		fmt.Printf("limit market order failed with err: %s\n", req.ErrorResponse.GetError())
		return nil, errors.New(req.ErrorResponse.GetMessage())
	}

	order, err := c.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		return nil, err
	} else if order.Status == nil {
		return order, errors.New("unknown order issue")
	}

	status := string(*order.Status)

	if status == "FILLED" || status == "OPEN" {
		return order, nil
	}

	return order, fmt.Errorf("order failed with status: '%s' and msg: '%s'", status, utils.StringPtrToString(order.CancelMessage)+" "+utils.StringPtrToString(order.RejectMessage))
}

func (c *apiclient) VerifyMarketOrderCompletion(ctx context.Context, orderID string, timeout time.Time) error {
	if time.Now().After(timeout) {
		err := fmt.Errorf("market order '%s' has timed out", orderID)
		c.log("%s", err.Error())
		return err
	}

	orderRes, err := c.client.GetOrder(orderID)
	if err != nil {
		log.Printf("\n")
		return err
	}

	order := orderRes.GetOrder()

	status := string(*order.Status)

	/* Coinbase order status'
	OPEN OrderStatus = "OPEN"
	FILLED OrderStatus = "FILLED"
	CANCELLED OrderStatus = "CANCELLED"
	EXPIRED OrderStatus = "EXPIRED"
	FAILED OrderStatus = "FAILED"
	*/
	switch status {
	/* These will be most if not all of the cases */
	case "OPEN":
		if !c.hasMentionedOrderWaiting {
			c.log("sleeping because market order has not completed yet...")
			c.hasMentionedOrderWaiting = true
		}
		time.Sleep(time.Second * 5)
		return c.VerifyMarketOrderCompletion(ctx, orderID, timeout)
	case "FILLED":
		c.hasMentionedOrderWaiting = false
		c.log("market order '%s' has completed", orderID)
		return nil
	/* These probably will never happen */
	case "CANCELLED", "FAILED":
		c.hasMentionedOrderWaiting = false
		return errors.New(order.GetCancelMessage() + " " + order.GetRejectReason())
	// This should theorectically never happen because we are cancelling fairly quickly on our own
	case "EXPIRED":
		c.hasMentionedOrderWaiting = false
		log.Println("order has expired and is being cancelled")
		if err = c.CancelOrders(ctx, orderID); err != nil {
			log.Println("unable to cancel order: ", err.Error())
		}

		return errors.New("order expired")
	}

	// This will never happen because coinbase doesn't have any other messages
	c.hasMentionedOrderWaiting = false
	return errors.New("unknown issue with the order: " + status)
}

func (c *apiclient) GetOrder(ctx context.Context, orderID string) (*cbadvmodel.Order, error) {
	res, err := c.client.GetOrder(orderID)
	if err != nil {
		return nil, err
	}

	return res.Order, nil
}

func (c *apiclient) GetOpenOrdersByProductIDAndSide(ctx context.Context, productID string, side cbadvmodel.OrderSide) ([]cbadvmodel.Order, error) {
	res, err := c.client.ListOrders(&cbadvclient.ListOrdersParams{
		ProductId:   productID,
		Limit:       250, // This should never even come close to being reached
		OrderStatus: []string{"OPEN"},
		OrderSide:   side,
	})
	if err != nil {
		return nil, err
	}

	return res.Orders, nil
}

func (c *apiclient) GetOrderFills(ctx context.Context, orderID, productID string) ([]cbadvmodel.OrderFill, error) {
	res, err := c.client.ListFills(&cbadvclient.ListFillsParams{
		OrderId:                orderID,
		ProductId:              productID,
		Limit:                  250,
		StartSequenceTimestamp: time.Now().Add(time.Minute * 5),
		EndSequenceTimestamp:   time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return res.Fills, nil
}

func (c *apiclient) CancelOrders(ctx context.Context, orderIds ...string) (err error) {
	res, err := c.client.CancelOrders(orderIds)
	if err != nil {
		return err
	}

	results := res.GetResults()
	if len(results) == 0 {
		return errors.New("no results returned from server for cancel order")
	}

	result := results[0]

	if result.GetSuccess() {
		return nil
	}

	return errors.New(result.GetFailureReason())
}

func (c *apiclient) CancelExistingOrders(
	ctx context.Context, id string, productID string, orderType model.OrderType,
) (err error) {
	ordersRes, err := c.client.ListOrders(&cbadvclient.ListOrdersParams{
		ProductId:          productID,
		StartDate:          time.Now().Add(time.Hour * -24),
		EndDate:            time.Now(),
		UserNativeCurrency: "USD",
		OrderType:          orderType,
		OrderSide:          model.UNKNOWN_ORDER_SIDE,
	})
	if err != nil {
		return err
	}

	for _, order := range ordersRes.Orders {
		if strings.Contains(order.GetClientOrderId(), id) {
			if err = c.CancelOrders(ctx, order.GetClientOrderId()); err != nil {
				return err
			}
		}
	}

	return nil
}

// log essentially force a new line
func (c *apiclient) log(format string, v ...any) {
	if c.debug {
		log.Printf(format+"\n", v...)
	}
}

/*
NOTES:

Validate Order:
(*model.Order)(0xc0004a8460)({
 OrderId: (*string)(0xc000258080)((len=36) "4c594cf2-31bf-4f7a-aa89-50493ab1c9e8"),
 ProductId: (*string)(0xc000258090)((len=7) "YFI-BTC"),
 UserId: (*string)(0xc0002580a0)((len=36) "825af154-d5d1-520e-b060-6cae5d5061e0"),
 OrderConfiguration: (*model.OutputOrderConfiguration)(0xc000512030)({
  MarketMarketIoc: (*model.OutputOrderConfigurationMarketMarketIoc)(<nil>),
  LimitLimitGtc: (*model.OutputOrderConfigurationLimitLimitGtc)(0xc00044a090)({
   BaseSize: (*float64)(0xc00002e480)(0.035252),
   LimitPrice: (*float64)(0xc00002e4b0)(0.4012),
   PostOnly: (*bool)(0xc00002e4be)(true)
  }),
  LimitLimitGtd: (*model.OutputOrderConfigurationLimitLimitGtd)(<nil>),
  StopLimitStopLimitGtc: (*model.OutputOrderConfigurationStopLimitStopLimitGtc)(<nil>),
  StopLimitStopLimitGtd: (*model.OutputOrderConfigurationStopLimitStopLimitGtd)(<nil>)
 }),
 Side: (*string)(0xc0002580d0)((len=3) "BUY"),
 ClientOrderId: (*string)(0xc0002580e0)((len=45) "create-market-order-BUY-YFI-BTC-1677204473976"),
 Status: (*model.OrderStatus)(0xc0002580f0)((len=4) "OPEN"),
 TimeInForce: (*string)(0xc000258120)((len=20) "GOOD_UNTIL_CANCELLED"),
 CreatedTime: (*string)(0xc000258130)((len=27) "2023-02-24T02:07:54.189564Z"),
 CompletionPercentage: (*float64)(0xc00002e4f0)(0),
 FilledSize: (*float64)(0xc00002e500)(0),
 AverageFilledPrice: (*float64)(0xc00002e520)(0),
 Fee: (*string)(0xc0002581c0)(""),
 NumberOfFills: (*float64)(0xc00002e530)(0),
 FilledValue: (*float64)(0xc00002e550)(0),
 PendingCancel: (*bool)(0xc00002e558)(false),
 SizeInQuote: (*bool)(0xc00002e559)(false),
 TotalFees: (*float64)(0xc00002e568)(0),
 SizeInclusiveOfFees: (*bool)(0xc00002e580)(false),
 TotalValueAfterFees: (*float64)(0xc00002e680)(0),
 TriggerStatus: (*string)(0xc000258210)((len=18) "INVALID_ORDER_TYPE"),
 OrderType: (*string)(0xc000258220)((len=5) "LIMIT"),
 RejectReason: (*string)(0xc000258230)((len=25) "REJECT_REASON_UNSPECIFIED"),
 Settled: (*bool)(0xc00002e68d)(false),
 ProductType: (*string)(0xc000258240)((len=4) "SPOT"),
 RejectMessage: (*string)(0xc000258250)(""),
 CancelMessage: (*string)(0xc000258260)("")
})

Invalid Order:
(*model.Order)(0xc0003fe000)({
 OrderId: (*string)(0xc0001143f0)((len=36) "b4164630-6c56-4380-bcb5-8aa692954361"),
 ProductId: (*string)(0xc000114400)((len=7) "YFI-BTC"),
 UserId: (*string)(0xc000114410)((len=36) "825af154-d5d1-520e-b060-6cae5d5061e0"),
 OrderConfiguration: (*model.OutputOrderConfiguration)(0xc000482960)({
  MarketMarketIoc: (*model.OutputOrderConfigurationMarketMarketIoc)(<nil>),
  LimitLimitGtc: (*model.OutputOrderConfigurationLimitLimitGtc)(0xc000598360)({
   BaseSize: (*float64)(0xc00011acf8)(0.035953),
   LimitPrice: (*float64)(0xc00011ad18)(0.3994),
   PostOnly: (*bool)(0xc00011ad26)(true)
  }),
  LimitLimitGtd: (*model.OutputOrderConfigurationLimitLimitGtd)(<nil>),
  StopLimitStopLimitGtc: (*model.OutputOrderConfigurationStopLimitStopLimitGtc)(<nil>),
  StopLimitStopLimitGtd: (*model.OutputOrderConfigurationStopLimitStopLimitGtd)(<nil>)
 }),
 Side: (*string)(0xc000114440)((len=3) "BUY"),
 ClientOrderId: (*string)(0xc000114450)((len=23) "create-market-order-BUY"),
 Status: (*model.OrderStatus)(0xc000114460)((len=9) "CANCELLED"),
 TimeInForce: (*string)(0xc000114480)((len=20) "GOOD_UNTIL_CANCELLED"),
 CreatedTime: (*string)(0xc000114490)((len=27) "2023-02-24T01:54:17.875968Z"),
 CompletionPercentage: (*float64)(0xc00011ad48)(0),
 FilledSize: (*float64)(0xc00011ad58)(0),
 AverageFilledPrice: (*float64)(0xc00011ad68)(0),
 Fee: (*string)(0xc0001144d0)(""),
 NumberOfFills: (*float64)(0xc00011ad78)(0),
 FilledValue: (*float64)(0xc00011ad88)(0),
 PendingCancel: (*bool)(0xc00011ad90)(false),
 SizeInQuote: (*bool)(0xc00011ad91)(false),
 TotalFees: (*float64)(0xc00011ada0)(0),
 SizeInclusiveOfFees: (*bool)(0xc00011ada8)(false),
 TotalValueAfterFees: (*float64)(0xc00011adb8)(0),
 TriggerStatus: (*string)(0xc000114520)((len=18) "INVALID_ORDER_TYPE"),
 OrderType: (*string)(0xc000114530)((len=5) "LIMIT"),
 RejectReason: (*string)(0xc000114540)((len=25) "REJECT_REASON_UNSPECIFIED"),
 Settled: (*bool)(0xc00011adc5)(false),
 ProductType: (*string)(0xc000114550)((len=4) "SPOT"),
 RejectMessage: (*string)(0xc000114560)(""),
 CancelMessage: (*string)(0xc000114570)((len=21) "User requested cancel")
})

(*cbadvmodel.Account)(0xc00041a180)({
  Uuid: (*string)(0xc0003dc9d0)((len=36) "f0a4ecb7-933c-53fc-9f4b-3c8fb07e42ab"),
  Name: (*string)(0xc0003dc9e0)((len=10) "BTC Wallet"),
  Currency: (*string)(0xc0003dc9f0)((len=3) "BTC"),
  AvailableBalance: (*cbadvmodel.AccountAvailableBalance)(0xc0003dca00)({
   Value: (*float64)(0xc00044abf0)(0),
   Currency: (*string)(0xc0003dca20)((len=3) "BTC")
  }),
  Default: (*bool)(0xc00044abfb)(true),
  Active: (*bool)(0xc00044abfc)(true),
  CreatedAt: (*string)(0xc0003dca30)((len=24) "2017-08-23T02:51:11.849Z"),
  UpdatedAt: (*string)(0xc0003dca40)((len=24) "2022-10-31T17:20:02.612Z"),
  DeletedAt: (*string)(<nil>),
  Type: (*string)(0xc0003dca50)((len=19) "ACCOUNT_TYPE_CRYPTO"),
  Ready: (*bool)(0xc00044abfd)(true),
  Hold: (*cbadvmodel.AccountAvailableBalance)(0xc0003dca60)({
   Value: (*float64)(0xc00044ac00)(0),
   Currency: (*string)(0xc0003dca80)((len=3) "BTC")
  })
 })

GetProduct:
(*model.GetProductResponse)(0xc0001d6000)({
 ProductId: (*string)(0xc000524290)((len=7) "LTC-BTC"),
 Price: (*float64)(0xc0003e8838)(0.00396),
 PricePercentageChange24h: (*float64)(0xc0003e8848)(0.55865921787709),
 Volume24h: (*float64)(0xc0003e88b0)(2344.85679773),
 VolumePercentageChange24h: (*float64)(0xc0003e88b8)(-59.0292876350855),
 BaseIncrement: (*float64)(0xc0003e88f0)(1e-08),
 QuoteIncrement: (*float64)(0xc0003e8918)(1e-06),
 QuoteMinSize: (*float64)(0xc0003e8938)(1.6e-05),
 QuoteMaxSize: (*float64)(0xc0003e8958)(200),
 BaseMinSize: (*float64)(0xc0003e8978)(0.0043),
 BaseMaxSize: (*float64)(0xc0003e89a8)(13000),
 BaseName: (*string)(0xc000524340)((len=8) "Litecoin"),
 QuoteName: (*string)(0xc000524350)((len=7) "Bitcoin"),
 Watched: (*bool)(0xc0003e89c7)(false),
 IsDisabled: (*bool)(0xc0003e89c8)(false),
 New: (*bool)(0xc0003e89c9)(false),
 Status: (*string)(0xc000524360)((len=6) "online"),
 CancelOnly: (*bool)(0xc0003e89d0)(false),
 LimitOnly: (*bool)(0xc0003e89d1)(false),
 PostOnly: (*bool)(0xc0003e89d2)(false),
 TradingDisabled: (*bool)(0xc0003e89d3)(false),
 AuctionMode: (*bool)(0xc0003e89d4)(false),
 ProductType: (*string)(0xc000524370)((len=4) "SPOT"),
 QuoteCurrencyId: (*string)(0xc000524380)((len=3) "BTC"),
 BaseCurrencyId: (*string)(0xc000524390)((len=3) "LTC"),
 FcmTradingSessionDetails: (*string)(<nil>),
 MidMarketPrice: (*string)(0xc0005243a0)(""),
 BaseDisplaySymbol: (*string)(0xc0005243b0)((len=3) "LTC"),
 QuoteDisplaySymbol: (*string)(0xc0005243c0)((len=3) "BTC")
})

*/
