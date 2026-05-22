package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	mu        sync.RWMutex
	users     []User
	customers []Customer
	services  []Service
	orders    []Order
	nextID    int
}

func NewStore() *Store {
	now := time.Date(2025, time.May, 18, 10, 30, 0, 0, time.Local)

	customers := []Customer{
		{ID: 1, Name: "Budi Santoso", Phone: "0812-3456-7890", CreatedAt: now.AddDate(0, 0, -22)},
		{ID: 2, Name: "Rina Amelia", Phone: "0813-9876-5432", CreatedAt: now.AddDate(0, 0, -18)},
		{ID: 3, Name: "Doni Pratama", Phone: "0812-1111-2222", CreatedAt: now.AddDate(0, 0, -16)},
		{ID: 4, Name: "Andi Wijaya", Phone: "0821-2222-3333", CreatedAt: now.AddDate(0, 0, -12)},
		{ID: 5, Name: "Siti Nurhaliza", Phone: "0812-5555-6677", CreatedAt: now.AddDate(0, 0, -9)},
		{ID: 6, Name: "Fajar Setiawan", Phone: "0813-4444-8888", CreatedAt: now.AddDate(0, 0, -7)},
	}

	services := []Service{
		{ID: 1, Name: "Fast Clean", Price: 45000, Duration: "1 hari", Active: true},
		{ID: 2, Name: "Deep Clean", Price: 65000, Duration: "2 hari", Active: true},
		{ID: 3, Name: "Unyellowing", Price: 85000, Duration: "2 hari", Active: true},
		{ID: 4, Name: "Repaint", Price: 90000, Duration: "2 hari", Active: true},
		{ID: 5, Name: "Repair", Price: 55000, Duration: "3 hari", Active: true},
	}

	orders := []Order{
		seedOrder(1, "INV-250518-0012", customers[0], StatusCompleted, PaymentPaid, "QRIS", now, now.AddDate(0, 0, 2), []OrderItem{{ID: 1, ServiceID: 2, Service: "Deep Clean", ShoeName: "Nike Air Force", Qty: 1, Price: 65000}}, "Sepatu putih, noda bagian sol."),
		seedOrder(2, "INV-250518-0011", customers[1], StatusCleaning, PaymentPartial, "Transfer", now.Add(45*time.Minute), now.AddDate(0, 0, 2), []OrderItem{{ID: 2, ServiceID: 3, Service: "Unyellowing", ShoeName: "Vans Old Skool", Qty: 1, Price: 85000}}, "Prioritas bagian midsole."),
		seedOrder(3, "INV-250518-0010", customers[2], StatusAccepted, PaymentUnpaid, "Cash", now.Add(-45*time.Minute), now.AddDate(0, 0, 1), []OrderItem{{ID: 3, ServiceID: 1, Service: "Fast Clean", ShoeName: "Adidas Campus", Qty: 1, Price: 45000}}, ""),
		seedOrder(4, "INV-250517-0009", customers[3], StatusPickedUp, PaymentPaid, "Cash", now.AddDate(0, 0, -1).Add(230*time.Minute), now.AddDate(0, 0, 2), []OrderItem{{ID: 4, ServiceID: 2, Service: "Deep Clean", ShoeName: "New Balance 530", Qty: 1, Price: 65000}, {ID: 5, ServiceID: 5, Service: "Repair", ShoeName: "New Balance 530", Qty: 1, Price: 55000}}, "Lem bagian depan."),
		seedOrder(5, "INV-250517-0008", customers[4], StatusWaitingPickup, PaymentPaid, "QRIS", now.AddDate(0, 0, -1).Add(-25*time.Minute), now.AddDate(0, 0, 1), []OrderItem{{ID: 6, ServiceID: 2, Service: "Deep Clean", ShoeName: "Asics Gel", Qty: 1, Price: 70000}}, "Hubungi sebelum pickup."),
		seedOrder(6, "INV-250516-0007", customers[5], StatusCancelled, PaymentUnpaid, "Cash", now.AddDate(0, 0, -2).Add(-60*time.Minute), now, []OrderItem{{ID: 7, ServiceID: 4, Service: "Repaint", ShoeName: "Jordan 1", Qty: 1, Price: 90000}}, "Customer batal repaint."),
	}

	return &Store{
		users:     []User{{ID: 1, Name: "Zolix Admin", Email: "admin@zolix.test", PasswordHash: hashPassword("admin123"), Role: RoleSuperAdmin}},
		customers: customers,
		services:  services,
		orders:    orders,
		nextID:    7,
	}
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func seedOrder(id int, invoice string, customer Customer, status Status, payment PaymentStatus, method string, createdAt, doneAt time.Time, items []OrderItem, notes string) Order {
	total := 0
	for _, item := range items {
		total += item.Price * item.Qty
	}
	return Order{
		ID:              id,
		InvoiceNumber:   invoice,
		CustomerID:      customer.ID,
		CustomerName:    customer.Name,
		CustomerPhone:   customer.Phone,
		Status:          status,
		TotalPrice:      total,
		PaymentStatus:   payment,
		PaymentMethod:   method,
		Notes:           notes,
		CreatedAt:       createdAt,
		EstimatedDoneAt: doneAt,
		Items:           items,
		Media: []Media{
			{ID: id*10 + 1, OrderID: id, Type: "before", URL: "/assets/image1.png"},
			{ID: id*10 + 2, OrderID: id, Type: "after", URL: "/assets/nota_order.png"},
		},
		Timeline: buildTimeline(status, createdAt),
	}
}

func buildTimeline(status Status, start time.Time) []Timeline {
	steps := []Status{StatusAccepted, StatusCleaning, StatusDrying, StatusFinishing, StatusReadyPickup, StatusCompleted, StatusPickedUp}
	order := map[Status]int{}
	for i, step := range steps {
		order[step] = i
	}
	order[StatusWaitingPickup] = order[StatusReadyPickup]
	current, ok := order[status]
	if !ok {
		current = -1
	}
	timeline := make([]Timeline, 0, len(steps))
	for i, step := range steps {
		done := i <= current
		var at time.Time
		if done {
			at = start.Add(time.Duration(i) * 4 * time.Hour)
		}
		timeline = append(timeline, Timeline{Label: string(step), Done: done, Time: at})
	}
	return timeline
}

func (s *Store) Dashboard() Dashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statusCounts := map[Status]int{}
	serviceAnalytics := map[string]int{}
	dailyRevenue := 0
	monthlyRevenue := 0
	active := 0
	completed := 0
	waiting := 0
	today := time.Date(2025, time.May, 18, 0, 0, 0, 0, time.Local)

	for _, order := range s.orders {
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

	recent := append([]Order(nil), s.orders...)
	sort.Slice(recent, func(i, j int) bool { return recent[i].CreatedAt.After(recent[j].CreatedAt) })
	if len(recent) > 6 {
		recent = recent[:6]
	}

	return Dashboard{
		TotalToday:       countToday(s.orders, today),
		ActiveOrders:     active,
		CompletedOrders:  completed,
		WaitingPickup:    waiting,
		DailyRevenue:     dailyRevenue,
		MonthlyRevenue:   monthlyRevenue,
		StatusCounts:     statusCounts,
		ServiceAnalytics: serviceAnalytics,
		RecentOrders:     recent,
		Activities: []string{
			"Invoice INV-250518-0012 terkirim via WhatsApp",
			"Rina Amelia masuk tahap cleaning",
			"Pickup reminder dikirim ke Siti Nurhaliza",
		},
	}
}

func countToday(orders []Order, today time.Time) int {
	total := 0
	for _, order := range orders {
		if order.CreatedAt.After(today) {
			total++
		}
	}
	return total
}

func (s *Store) Orders(query, status string) []Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	status = strings.TrimSpace(status)
	orders := make([]Order, 0, len(s.orders))
	for _, order := range s.orders {
		if status != "" && status != "Semua" && string(order.Status) != status {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(order.InvoiceNumber + " " + order.CustomerName + " " + order.CustomerPhone + " " + order.Items[0].Service)
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		orders = append(orders, order)
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].CreatedAt.After(orders[j].CreatedAt) })
	return orders
}

func (s *Store) Order(id int) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, order := range s.orders {
		if order.ID == id {
			return order, true
		}
	}
	return Order{}, false
}

func (s *Store) OrderByInvoice(invoice string) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, order := range s.orders {
		if strings.EqualFold(order.InvoiceNumber, strings.TrimSpace(invoice)) {
			return order, true
		}
	}
	return Order{}, false
}

func (s *Store) CreateOrder(input Order) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(input.CustomerName) == "" || len(input.Items) == 0 {
		return Order{}, errors.New("customer and item are required")
	}

	input.ID = s.nextID
	input.InvoiceNumber = fmt.Sprintf("INV-250518-%04d", s.nextID)
	input.CreatedAt = time.Date(2025, time.May, 18, 13, 20, 0, 0, time.Local)
	input.Status = StatusAccepted
	input.Timeline = buildTimeline(input.Status, input.CreatedAt)
	input.CustomerID = s.nextID
	input.TotalPrice = 0
	for i := range input.Items {
		input.Items[i].ID = s.nextID*10 + i
		input.TotalPrice += input.Items[i].Price * input.Items[i].Qty
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = "Cash"
	}
	if input.PaymentStatus == "" {
		input.PaymentStatus = PaymentUnpaid
	}
	if input.EstimatedDoneAt.IsZero() {
		input.EstimatedDoneAt = input.CreatedAt.AddDate(0, 0, 2)
	}

	s.orders = append(s.orders, input)
	s.customers = append(s.customers, Customer{ID: input.CustomerID, Name: input.CustomerName, Phone: input.CustomerPhone, CreatedAt: input.CreatedAt})
	s.nextID++
	return input, nil
}

func (s *Store) UpdateOrder(id int, input UpdateOrderRequest) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orders {
		if s.orders[i].ID != id {
			continue
		}
		if strings.TrimSpace(input.CustomerName) != "" {
			s.orders[i].CustomerName = input.CustomerName
		}
		if strings.TrimSpace(input.CustomerPhone) != "" {
			s.orders[i].CustomerPhone = input.CustomerPhone
		}
		if input.Status != "" {
			s.orders[i].Status = input.Status
			s.orders[i].Timeline = buildTimeline(input.Status, s.orders[i].CreatedAt)
		}
		if input.PaymentStatus != "" {
			s.orders[i].PaymentStatus = input.PaymentStatus
		}
		if strings.TrimSpace(input.PaymentMethod) != "" {
			s.orders[i].PaymentMethod = input.PaymentMethod
		}
		if !input.EstimatedDoneAt.IsZero() {
			s.orders[i].EstimatedDoneAt = input.EstimatedDoneAt
		}
		s.orders[i].Notes = input.Notes
		if len(input.Items) > 0 {
			total := 0
			for itemIndex := range input.Items {
				input.Items[itemIndex].ID = id*10 + itemIndex
				total += input.Items[itemIndex].Price * input.Items[itemIndex].Qty
			}
			s.orders[i].Items = input.Items
			s.orders[i].TotalPrice = total
		}
		return s.orders[i], nil
	}
	return Order{}, errors.New("order not found")
}

func (s *Store) DeleteOrder(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orders {
		if s.orders[i].ID == id {
			s.orders = append(s.orders[:i], s.orders[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) AddMedia(orderID int, mediaType, url string) (Media, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.orders {
		if s.orders[i].ID != orderID {
			continue
		}
		media := Media{
			ID:      orderID*100 + len(s.orders[i].Media) + 1,
			OrderID: orderID,
			Type:    mediaType,
			URL:     url,
		}
		s.orders[i].Media = append(s.orders[i].Media, media)
		return media, nil
	}
	return Media{}, errors.New("order not found")
}

func (s *Store) UpdateOrderPayment(id int, update PaymentUpdate) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orders {
		if s.orders[i].ID != id {
			continue
		}
		if update.PaymentStatus != "" {
			s.orders[i].PaymentStatus = update.PaymentStatus
		}
		if strings.TrimSpace(update.PaymentMethod) != "" {
			s.orders[i].PaymentMethod = update.PaymentMethod
		}
		s.orders[i].PaymentProvider = update.PaymentProvider
		s.orders[i].PaymentReference = update.PaymentReference
		s.orders[i].PaymentExternalOrderID = update.PaymentExternalOrderID
		s.orders[i].PaymentQRString = update.PaymentQRString
		s.orders[i].PaymentQRURL = update.PaymentQRURL
		s.orders[i].PaymentExpiryTime = update.PaymentExpiryTime
		s.orders[i].PaymentUpdatedAt = update.PaymentUpdatedAt
		return s.orders[i], nil
	}
	return Order{}, errors.New("order not found")
}

func (s *Store) OrderByPaymentReference(reference string) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	reference = strings.TrimSpace(reference)
	for _, order := range s.orders {
		if reference != "" && order.PaymentReference == reference {
			return order, true
		}
	}
	return Order{}, false
}

func (s *Store) Customers() []Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Customer(nil), s.customers...)
}

func (s *Store) Services() []Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Service(nil), s.services...)
}

func (s *Store) Users() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]User(nil), s.users...)
}

func (s *Store) Authenticate(email, password string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if strings.EqualFold(user.Email, strings.TrimSpace(email)) && checkPassword(user.PasswordHash, password) {
			return user, true
		}
	}
	return User{}, false
}
