package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"readermicroservice/internal/config"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
)

// Структуры данных для заказа
type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

const path = "readermicroservice/configs/main.yml"

func main() {
	var (
		n        = flag.Int("n", 100, "how many orders to generate and send")
		interval = flag.Duration("interval", 50*time.Millisecond, "delay between messages (e.g. 10ms, 200ms, 1s)")
		seed     = flag.Int64("seed", time.Now().UnixNano(), "random seed (use fixed for reproducible runs)")
	)
	flag.Parse()

	kafkaConfig, err := config.LoadKafkaConfig(path)
	if err != nil {
		log.Fatal("Error loading Kafka config: ", err)
	}

	ctx := context.Background()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaConfig.Brokers...),
		Topic:        kafkaConfig.Topic,
		BatchSize:    1,
		Async:        false,
		RequiredAcks: kafka.RequireAll,
	}
	defer writer.Close()

	// Инициализация генераторов
	gofakeit.Seed(*seed)
	rng := rand.New(rand.NewSource(*seed))

	log.Printf("Sending %d messages to topic=%s brokers=%v interval=%s seed=%d",
		*n, kafkaConfig.Topic, kafkaConfig.Brokers, interval.String(), *seed,
	)

	for i := 0; i < *n; i++ {
		order := generateOrder(rng)

		b, err := json.Marshal(order)
		if err != nil {
			log.Printf("marshal error: %v", err)
			continue
		}

		err = writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(order.OrderUID), // полезно для партиционирования/дедупа
			Value: b,
			Time:  time.Now(),
		})
		if err != nil {
			log.Printf("send error (order_uid=%s): %v", order.OrderUID, err)
			continue
		}

		log.Printf("sent %d/%d order_uid=%s items=%d", i+1, *n, order.OrderUID, len(order.Items))

		if *interval > 0 {
			time.Sleep(*interval)
		}
	}

	log.Println("Done.")
}

func generateOrder(rng *rand.Rand) Order {
	uid := gofakeit.UUID()
	track := fmt.Sprintf("WB-%s", gofakeit.LetterN(12))

	itemsCount := rng.Intn(5) + 1 // 1..5
	items := make([]Item, 0, itemsCount)

	goodsTotal := 0
	for i := 0; i < itemsCount; i++ {
		price := rng.Intn(5000) + 100       // 100..5100
		sale := rng.Intn(61)                // 0..60
		total := price * (100 - sale) / 100 // со скидкой
		goodsTotal += total

		items = append(items, Item{
			ChrtID:      rng.Intn(9_999_999),
			TrackNumber: track,
			Price:       price,
			Rid:         gofakeit.UUID(),
			Name:        gofakeit.Word(),
			Sale:        sale,
			Size:        fmt.Sprintf("%d", rng.Intn(5)), // простая заглушка
			TotalPrice:  total,
			NmID:        rng.Intn(9_999_999),
			Brand:       gofakeit.Company(),
			Status:      200 + rng.Intn(30),
		})
	}

	deliveryCost := rng.Intn(2000) // 0..1999
	amount := goodsTotal + deliveryCost

	created := time.Now().Add(-time.Duration(rng.Intn(3600*24*30)) * time.Second) // до ~30 дней назад

	return Order{
		OrderUID:    uid,
		TrackNumber: track,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Address,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: Payment{
			Transaction:  uid,
			RequestID:    "",
			Currency:     gofakeit.RandomString([]string{"RUB", "USD", "EUR"}),
			Provider:     "wbpay",
			Amount:       amount,
			PaymentDt:    created.Unix(),
			Bank:         gofakeit.RandomString([]string{"alpha", "tinkoff", "sber", "vtb"}),
			DeliveryCost: deliveryCost,
			GoodsTotal:   goodsTotal,
			CustomFee:    0,
		},
		Items:             items,
		Locale:            gofakeit.RandomString([]string{"ru", "en"}),
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("user-%d", rng.Intn(10_000)),
		DeliveryService:   gofakeit.RandomString([]string{"meest", "cdek", "dhl", "dpd"}),
		Shardkey:          fmt.Sprintf("%d", rng.Intn(10)),
		SmID:              rng.Intn(1000),
		DateCreated:       created.UTC(),
		OofShard:          fmt.Sprintf("%d", rng.Intn(5)),
	}
}

func seedToUint64(s int64) uint64 {
	if s < 0 {
		return uint64(-s)
	}
	return uint64(s)
}
