package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	DSN string
}

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(config PostgresConfig) (*PostgresStore, error) {
	if strings.TrimSpace(config.DSN) == "" {
		return nil, errors.New("DATABASE_URL is required when STORE=postgres")
	}
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	store := &PostgresStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}
	if err := store.seed(); err != nil {
		return nil, err
	}
	if err := store.ensureAdminPassword(); err != nil {
		return nil, err
	}
	if err := store.resetSequences(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) migrate() error {
	return s.db.AutoMigrate(&User{}, &Customer{}, &Service{}, &Order{}, &OrderItem{}, &Media{}, &Timeline{})
}

func (s *PostgresStore) seed() error {
	var count int64
	if err := s.db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	memory := NewStore()
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&memory.users).Error; err != nil {
			return err
		}
		if err := tx.Create(&memory.customers).Error; err != nil {
			return err
		}
		if err := tx.Create(&memory.services).Error; err != nil {
			return err
		}
		for _, order := range memory.orders {
			for i := range order.Items {
				order.Items[i].OrderID = order.ID
			}
			for i := range order.Media {
				order.Media[i].OrderID = order.ID
			}
			for i := range order.Timeline {
				order.Timeline[i].OrderID = order.ID
			}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *PostgresStore) ensureAdminPassword() error {
	return s.db.Model(&User{}).
		Where("email = ? AND (password_hash = ? OR password_hash IS NULL)", "admin@zolix.test", "").
		Update("password_hash", hashPassword("admin123")).Error
}

func (s *PostgresStore) resetSequences() error {
	tables := []string{"users", "customers", "services", "orders", "order_items", "media", "timelines"}
	for _, table := range tables {
		sql := fmt.Sprintf(
			"SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM %s), 1), true)",
			table,
			table,
		)
		if err := s.db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) Dashboard() Dashboard {
	orders := s.Orders("", "")

	statusCounts := map[Status]int{}
	serviceAnalytics := map[string]int{}
	dailyRevenue := 0
	monthlyRevenue := 0
	active := 0
	completed := 0
	waiting := 0
	today := time.Date(2025, time.May, 18, 0, 0, 0, 0, time.Local)

	for _, order := range orders {
		statusCounts[order.Status]++
		if order.CreatedAt.After(today) && order.PaymentStatus == PaymentPaid {
			dailyRevenue += order.TotalPrice
		}
		if order.PaymentStatus == PaymentPaid {
			monthlyRevenue += order.TotalPrice
		}
		switch order.Status {
		case StatusCompleted, StatusPickedUp:
			completed++
		case StatusWaitingPickup:
			waiting++
			active++
		case StatusCancelled:
		default:
			active++
		}
		for _, item := range order.Items {
			serviceAnalytics[item.Service] += item.Qty
		}
	}

	recent := append([]Order(nil), orders...)
	if len(recent) > 6 {
		recent = recent[:6]
	}

	return Dashboard{
		TotalToday:       countToday(orders, today),
		ActiveOrders:     active,
		CompletedOrders:  completed,
		WaitingPickup:    waiting,
		DailyRevenue:     dailyRevenue,
		MonthlyRevenue:   monthlyRevenue,
		StatusCounts:     statusCounts,
		ServiceAnalytics: serviceAnalytics,
		RecentOrders:     recent,
		Activities: []string{
			"PostgreSQL store aktif dan siap menerima order",
			"Auto migration berhasil dijalankan",
			"Seed layanan default tersedia",
		},
	}
}

func (s *PostgresStore) Orders(query, status string) []Order {
	var orders []Order
	db := s.db.Preload("Items").Preload("Media").Preload("Timeline")
	if status = strings.TrimSpace(status); status != "" && status != "Semua" {
		db = db.Where("status = ?", status)
	}
	if query = strings.ToLower(strings.TrimSpace(query)); query != "" {
		like := "%" + query + "%"
		db = db.Where("lower(invoice_number) LIKE ? OR lower(customer_name) LIKE ? OR lower(customer_phone) LIKE ?", like, like, like)
	}
	if err := db.Order("created_at desc").Find(&orders).Error; err != nil {
		return nil
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders
}

func (s *PostgresStore) Order(id int) (Order, bool) {
	var order Order
	err := s.db.Preload("Items").Preload("Media").Preload("Timeline").First(&order, id).Error
	if err != nil {
		return Order{}, false
	}
	return order, true
}

func (s *PostgresStore) OrderByInvoice(invoice string) (Order, bool) {
	var order Order
	err := s.db.Preload("Items").Preload("Media").Preload("Timeline").
		Where("invoice_number = ?", strings.TrimSpace(invoice)).
		First(&order).Error
	if err != nil {
		return Order{}, false
	}
	return order, true
}

func (s *PostgresStore) CreateOrder(input Order) (Order, error) {
	if strings.TrimSpace(input.CustomerName) == "" || len(input.Items) == 0 {
		return Order{}, errors.New("customer and item are required")
	}

	now := time.Now()
	if input.CreatedAt.IsZero() {
		input.CreatedAt = now
	}
	if input.EstimatedDoneAt.IsZero() {
		input.EstimatedDoneAt = input.CreatedAt.AddDate(0, 0, 2)
	}
	if input.Status == "" {
		input.Status = StatusAccepted
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = "Cash"
	}
	if input.PaymentStatus == "" {
		input.PaymentStatus = PaymentUnpaid
	}
	for _, item := range input.Items {
		input.TotalPrice += item.Price * item.Qty
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		customer := Customer{Name: input.CustomerName, Phone: input.CustomerPhone, CreatedAt: input.CreatedAt}
		if err := tx.Where("phone = ?", input.CustomerPhone).FirstOrCreate(&customer).Error; err != nil {
			return err
		}
		input.CustomerID = customer.ID
		input.Timeline = buildTimeline(input.Status, input.CreatedAt)
		orderRecord := input
		orderRecord.Items = nil
		orderRecord.Media = nil
		orderRecord.Timeline = nil
		if err := tx.Create(&orderRecord).Error; err != nil {
			return err
		}
		input.ID = orderRecord.ID
		input.InvoiceNumber = fmt.Sprintf("INV-%s-%04d", input.CreatedAt.Format("060102"), input.ID)
		if err := tx.Model(&orderRecord).Update("invoice_number", input.InvoiceNumber).Error; err != nil {
			return err
		}
		for i := range input.Items {
			input.Items[i].OrderID = input.ID
		}
		for i := range input.Timeline {
			input.Timeline[i].OrderID = input.ID
		}
		if err := tx.Create(&input.Items).Error; err != nil {
			return err
		}
		if err := tx.Create(&input.Timeline).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return Order{}, err
	}
	return input, nil
}

func (s *PostgresStore) UpdateOrder(id int, input UpdateOrderRequest) (Order, error) {
	order, ok := s.Order(id)
	if !ok {
		return Order{}, errors.New("order not found")
	}
	if strings.TrimSpace(input.CustomerName) != "" {
		order.CustomerName = input.CustomerName
	}
	if strings.TrimSpace(input.CustomerPhone) != "" {
		order.CustomerPhone = input.CustomerPhone
	}
	if input.Status != "" {
		order.Status = input.Status
		order.Timeline = buildTimeline(input.Status, order.CreatedAt)
	}
	if input.PaymentStatus != "" {
		order.PaymentStatus = input.PaymentStatus
	}
	if strings.TrimSpace(input.PaymentMethod) != "" {
		order.PaymentMethod = input.PaymentMethod
	}
	if !input.EstimatedDoneAt.IsZero() {
		order.EstimatedDoneAt = input.EstimatedDoneAt
	}
	order.Notes = input.Notes
	if len(input.Items) > 0 {
		order.Items = input.Items
		order.TotalPrice = 0
		for i := range order.Items {
			order.Items[i].OrderID = order.ID
			order.TotalPrice += order.Items[i].Price * order.Items[i].Qty
		}
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Order{}).Where("id = ?", id).Updates(map[string]any{
			"customer_name":     order.CustomerName,
			"customer_phone":    order.CustomerPhone,
			"status":            order.Status,
			"total_price":       order.TotalPrice,
			"payment_status":    order.PaymentStatus,
			"payment_method":    order.PaymentMethod,
			"notes":             order.Notes,
			"estimated_done_at": order.EstimatedDoneAt,
		}).Error; err != nil {
			return err
		}
		if len(input.Items) > 0 {
			if err := tx.Where("order_id = ?", id).Delete(&OrderItem{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&order.Items).Error; err != nil {
				return err
			}
		}
		if input.Status != "" {
			for i := range order.Timeline {
				order.Timeline[i].OrderID = id
				order.Timeline[i].ID = 0
			}
			if err := tx.Where("order_id = ?", id).Delete(&Timeline{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&order.Timeline).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return Order{}, err
	}
	updated, ok := s.Order(id)
	if !ok {
		return Order{}, errors.New("order not found")
	}
	return updated, nil
}

func (s *PostgresStore) DeleteOrder(id int) bool {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&Timeline{}).Error; err != nil {
			return err
		}
		if err := tx.Where("order_id = ?", id).Delete(&Media{}).Error; err != nil {
			return err
		}
		if err := tx.Where("order_id = ?", id).Delete(&OrderItem{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&Order{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("order not found")
		}
		return nil
	})
	return err == nil
}

func (s *PostgresStore) AddMedia(orderID int, mediaType, url string) (Media, error) {
	if _, ok := s.Order(orderID); !ok {
		return Media{}, errors.New("order not found")
	}
	media := Media{OrderID: orderID, Type: mediaType, URL: url}
	if err := s.db.Create(&media).Error; err != nil {
		return Media{}, err
	}
	return media, nil
}

func (s *PostgresStore) UpdateOrderPayment(id int, update PaymentUpdate) (Order, error) {
	values := map[string]any{
		"payment_provider":          update.PaymentProvider,
		"payment_reference":         update.PaymentReference,
		"payment_external_order_id": update.PaymentExternalOrderID,
		"payment_qr_string":         update.PaymentQRString,
		"payment_qr_url":            update.PaymentQRURL,
		"payment_expiry_time":       update.PaymentExpiryTime,
		"payment_updated_at":        update.PaymentUpdatedAt,
	}
	if update.PaymentStatus != "" {
		values["payment_status"] = update.PaymentStatus
	}
	if strings.TrimSpace(update.PaymentMethod) != "" {
		values["payment_method"] = update.PaymentMethod
	}
	result := s.db.Model(&Order{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		return Order{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Order{}, errors.New("order not found")
	}
	updated, ok := s.Order(id)
	if !ok {
		return Order{}, errors.New("order not found")
	}
	return updated, nil
}

func (s *PostgresStore) OrderByPaymentReference(reference string) (Order, bool) {
	var order Order
	err := s.db.Preload("Items").Preload("Media").Preload("Timeline").
		Where("payment_reference = ?", strings.TrimSpace(reference)).
		First(&order).Error
	if err != nil {
		return Order{}, false
	}
	return order, true
}

func (s *PostgresStore) Customers() []Customer {
	var customers []Customer
	if err := s.db.Order("created_at desc").Find(&customers).Error; err != nil {
		return nil
	}
	return customers
}

func (s *PostgresStore) Services() []Service {
	var services []Service
	if err := s.db.Order("id asc").Find(&services).Error; err != nil {
		return nil
	}
	return services
}

func (s *PostgresStore) Users() []User {
	var users []User
	if err := s.db.Order("id asc").Find(&users).Error; err != nil {
		return nil
	}
	return users
}

func (s *PostgresStore) Authenticate(email, password string) (User, bool) {
	var user User
	err := s.db.Where("lower(email) = lower(?)", strings.TrimSpace(email)).First(&user).Error
	if err != nil {
		return User{}, false
	}
	if !checkPassword(user.PasswordHash, password) {
		return User{}, false
	}
	return user, true
}
