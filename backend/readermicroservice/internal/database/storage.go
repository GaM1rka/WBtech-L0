package database

import (
	"database/sql"
	"fmt"
	"readermicroservice/configs"
	"readermicroservice/internal/kafka/models"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type Config struct {
	Port     string
	User     string
	Password string
	DBName   string
}

func New(cfg Config) (*DB, error) {
	connStr := fmt.Sprintf("port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		configs.RLogger.Println("Error while initializing new DB: ", err)
		return nil, err
	}
	if err = db.Ping(); err != nil {
		configs.RLogger.Println("Falied to ping DB: ", err)
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) CreateTables() error {
	configs.RLogger.Println("Creating Tables.")

	// Создание таблицы orders
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS orders (" +
		"order_uid VARCHAR(100) PRIMARY KEY," +
		"track_number VARCHAR(100)," +
		"entry VARCHAR(100)," +
		"locale VARCHAR(100)," +
		"internal_signature VARCHAR(100)," +
		"customer_id VARCHAR(100)," +
		"delivery_service VARCHAR(100)," +
		"shardkey VARCHAR(100)," +
		"sm_id INTEGER," +
		"date_created TIMESTAMP," +
		"oof_shard VARCHAR(100),")

	if err != nil {
		configs.RLogger.Println("Error while creating a table orders: ", err)
		return err
	}

	// Создание таблицы delivery
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS delivery (" +
		"id SERIAL PRIMARY KEY," +
		"order_uid VARCHAR(100) REFERENCES orders(order_uid)," +
		"name VARCHAR(100)," +
		"phone VARCHAR(100)," +
		"zip VARCHAR(100)," +
		"city VARCHAR(100)," +
		"address VARCHAR(255)," +
		"region VARCHAR(100)," +
		"email VARCHAR(100)" +
		")")

	if err != nil {
		configs.RLogger.Println("Error while creating a table delivery: ", err)
		return err
	}

	// Создание таблицы payments
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS payments (" +
		"id SERIAL PRIMARY KEY," +
		"order_uid VARCHAR(100) REFERENCES orders(order_uid)," +
		"transaction VARCHAR(100)," +
		"request_id VARCHAR(100)," +
		"currency VARCHAR(10)," +
		"provider VARCHAR(100)," +
		"amount INTEGER," +
		"payment_dt BIGINT," +
		"bank VARCHAR(100)," +
		"delivery_cost INTEGER," +
		"goods_total INTEGER," +
		"custom_fee INTEGER" +
		")")

	if err != nil {
		configs.RLogger.Println("Error while creating a table payments: ", err)
		return err
	}

	return nil
}

func (db *DB) Insert(data models.Order) error {
	_, err := db.Exec("INSERT INTO orders"+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		data.OrderUID, data.TrackNumber, data.Entry, data.Locale, data.InternalSignature,
		data.CustomerID, data.DeliveryService, data.Shardkey, data.SmID, data.DateCreated, data.OofShard)
	if err != nil {
		configs.RLogger.Println("Error while inserting data to orders table: ", err)
		return err
	}

	_, err = db.Exec("INSERT INTO delivery(order_uid, name, phone, zip, city, address, region, email)"+
		"VALUES ($1, $2, $3, $4, $5, $6, $7)", data.Delivery.Name, data.Delivery.Phone,
		data.Delivery.Zip, data.Delivery.City, data.Delivery.Address, data.Delivery.Region, data.Delivery.Email)
	if err != nil {
		configs.RLogger.Println("Error while inserting data to delivery table: ", err)
		return err
	}

	_, err = db.Exec("INSERT INTO payment(order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		data.OrderUID, data.Payment.Transaction, data.Payment.RequestID, data.Payment.Currency,
		data.Payment.Provider, data.Payment.Amount, data.Payment.PaymentDt, data.Payment.Bank,
		data.Payment.DeliveryCost, data.Payment.GoodsTotal, data.Payment.CustomFee)
	if err != nil {
		configs.RLogger.Println("Error while inserting data to payment table: ", err)
		return err
	}

	// Вставка items — по одному запросу на каждую позицию
	itemStmt := "INSERT INTO items(order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)"
	for _, it := range data.Items {
		_, err = db.Exec(itemStmt,
			data.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.Rid, it.Name,
			it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			configs.RLogger.Println("Error while inserting data to items table: ", err)
			return err
		}
	}

	return nil
}
