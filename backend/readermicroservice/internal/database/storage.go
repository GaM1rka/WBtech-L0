package database

import (
	"database/sql"
	"fmt"

	"readermicroservice/internal/config"
	"readermicroservice/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

const (
	path string = "readermicroservice/configs/main.yml"
)

func New(cfg models.Config) (*DB, error) {
	var DBConfig *config.DBConfig
	DBConfig, err := config.LoadDBConfig(path)
	if err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DBConfig.Host, DBConfig.Port, DBConfig.User, DBConfig.Password, DBConfig.Database)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		config.RLogger.Println("Error while initializing new DB: ", err)
		return nil, err
	}
	if err = db.Ping(); err != nil {
		config.RLogger.Println("Falied to ping DB: ", err)
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) CreateTables() error {
	config.RLogger.Println("Creating Tables.")

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
		"oof_shard VARCHAR(100)" +
		")")

	if err != nil {
		config.RLogger.Println("Error while creating a table orders: ", err)
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
		config.RLogger.Println("Error while creating a table delivery: ", err)
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
		config.RLogger.Println("Error while creating a table payments: ", err)
		return err
	}

	// Создание таблицы items
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS items (" +
		"id SERIAL PRIMARY KEY," +
		"order_uid VARCHAR(100) REFERENCES orders(order_uid)," +
		"chrt_id INTEGER," +
		"track_number VARCHAR(100)," +
		"price INTEGER," +
		"rid VARCHAR(100)," +
		"name VARCHAR(255)," +
		"sale INTEGER," +
		"size VARCHAR(50)," +
		"total_price INTEGER," +
		"nm_id INTEGER," +
		"brand VARCHAR(100)," +
		"status INTEGER" +
		")")

	if err != nil {
		config.RLogger.Println("Error while creating a table items: ", err)
		return err
	}

	config.RLogger.Println("Successfuly created tables!")
	return nil
}

func (db *DB) Insert(data models.Order) error {
	_, err := db.Exec("INSERT INTO orders "+
		"(order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		data.OrderUID, data.TrackNumber, data.Entry, data.Locale, data.InternalSignature,
		data.CustomerID, data.DeliveryService, data.Shardkey, data.SmID, data.DateCreated, data.OofShard)
	if err != nil {
		config.RLogger.Println("Error while inserting data to orders table: ", err)
		return err
	}

	_, err = db.Exec("INSERT INTO delivery(order_uid, name, phone, zip, city, address, region, email)"+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", data.OrderUID, data.Delivery.Name, data.Delivery.Phone,
		data.Delivery.Zip, data.Delivery.City, data.Delivery.Address, data.Delivery.Region, data.Delivery.Email)
	if err != nil {
		config.RLogger.Println("Error while inserting data to delivery table: ", err)
		return err
	}

	_, err = db.Exec("INSERT INTO payments(order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		data.OrderUID, data.Payment.Transaction, data.Payment.RequestID, data.Payment.Currency,
		data.Payment.Provider, data.Payment.Amount, data.Payment.PaymentDt, data.Payment.Bank,
		data.Payment.DeliveryCost, data.Payment.GoodsTotal, data.Payment.CustomFee)
	if err != nil {
		config.RLogger.Println("Error while inserting data to payment table: ", err)
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
			config.RLogger.Println("Error while inserting data to items table: ", err)
			return err
		}
	}

	return nil
}

func (db *DB) GetByUID(order_uid string) (*models.Order, error) {
	row := db.QueryRow("SELECT * FROM orders WHERE order_uid = $1", order_uid)
	var order models.Order
	err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey,
		&order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		config.RLogger.Println("Error scanning order: ", err)
		return nil, err
	}

	// Получаем данные доставки
	var delivery models.Delivery
	row = db.QueryRow("SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1", order_uid)
	err = row.Scan(&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
		&delivery.Address, &delivery.Region, &delivery.Email)
	if err != nil {
		config.RLogger.Println("Error getting delivery for order ", order.OrderUID, ": ", err)
		return nil, err
	}
	order.Delivery = delivery

	// Получаем данные оплаты
	var payment models.Payment
	row = db.QueryRow("SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE order_uid = $1", order_uid)
	err = row.Scan(&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider,
		&payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost,
		&payment.GoodsTotal, &payment.CustomFee)
	if err != nil {
		config.RLogger.Println("Error getting payment for order ", order.OrderUID, ": ", err)
		return nil, err
	}
	order.Payment = payment

	// Получаем товары
	itemRows, err := db.Query("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1", order_uid)
	if err != nil {
		config.RLogger.Println("Error getting items for order ", order.OrderUID, ": ", err)
		return nil, err
	}
	defer itemRows.Close()

	var items []models.Item
	for itemRows.Next() {
		var item models.Item
		err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			config.RLogger.Println("Error scanning item for order ", order.OrderUID, ": ", err)
			continue
		}
		items = append(items, item)
	}
	order.Items = items
	return &order, nil
}

func (db *DB) GetAll() ([]models.Order, error) {
	rows, err := db.Query("SELECT * FROM orders")
	if err != nil {
		config.RLogger.Println("Error while reading data from orders table: ", err)
		return nil, err
	}

	orders := make([]models.Order, 0)

	defer rows.Close()
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey,
			&order.SmID, &order.DateCreated, &order.OofShard)
		if err != nil {
			config.RLogger.Println("Error scanning order: ", err)
			continue
		}

		// Получаем данные доставки
		var delivery models.Delivery
		row := db.QueryRow("SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1", order.OrderUID)
		err = row.Scan(&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
			&delivery.Address, &delivery.Region, &delivery.Email)
		if err != nil {
			config.RLogger.Println("Error getting delivery for order ", order.OrderUID, ": ", err)
			continue
		}
		order.Delivery = delivery

		// Получаем данные оплаты
		var payment models.Payment
		row = db.QueryRow("SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE order_uid = $1", order.OrderUID)
		err = row.Scan(&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider,
			&payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost,
			&payment.GoodsTotal, &payment.CustomFee)
		if err != nil {
			config.RLogger.Println("Error getting payment for order ", order.OrderUID, ": ", err)
			continue
		}
		order.Payment = payment

		// Получаем товары
		itemRows, err := db.Query("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1", order.OrderUID)
		if err != nil {
			config.RLogger.Println("Error getting items for order ", order.OrderUID, ": ", err)
			continue
		}
		defer itemRows.Close()

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
				&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if err != nil {
				config.RLogger.Println("Error scanning item for order ", order.OrderUID, ": ", err)
				continue
			}
			items = append(items, item)
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		config.RLogger.Println("Error iterating orders: ", err)
		return nil, err
	}

	return orders, nil
}
