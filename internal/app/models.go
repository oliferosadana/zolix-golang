package app

import "time"

type Role string

const (
	RoleSuperAdmin Role = "Super Admin"
	RoleStaff      Role = "Staff"
	RoleCleaner    Role = "Cleaner"
)

type Status string

const (
	StatusPending       Status = "Pending"
	StatusAccepted      Status = "Diterima"
	StatusCleaning      Status = "Cleaning"
	StatusDrying        Status = "Drying"
	StatusFinishing     Status = "Finishing"
	StatusReadyPickup   Status = "Ready Pickup"
	StatusCompleted     Status = "Completed"
	StatusPickedUp      Status = "Diambil"
	StatusCancelled     Status = "Dibatalkan"
	StatusWaitingPickup Status = "Menunggu Diambil"
)

type PaymentStatus string

const (
	PaymentUnpaid  PaymentStatus = "Belum Bayar"
	PaymentPaid    PaymentStatus = "Lunas"
	PaymentPartial PaymentStatus = "DP"
)

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
}

type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type Service struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Duration string `json:"duration"`
	Active   bool   `json:"active"`
}

type OrderItem struct {
	ID        int    `json:"id"`
	OrderID   int    `json:"-"`
	ServiceID int    `json:"service_id"`
	Service   string `json:"service"`
	ShoeName  string `json:"shoe_name"`
	Qty       int    `json:"qty"`
	Price     int    `json:"price"`
}

type Media struct {
	ID      int    `json:"id"`
	OrderID int    `json:"order_id"`
	Type    string `json:"type"`
	URL     string `json:"url"`
}

type Order struct {
	ID                     int           `json:"id"`
	InvoiceNumber          string        `json:"invoice_number"`
	CustomerID             int           `json:"customer_id"`
	CustomerName           string        `json:"customer_name"`
	CustomerPhone          string        `json:"customer_phone"`
	Status                 Status        `json:"status"`
	TotalPrice             int           `json:"total_price"`
	PaymentStatus          PaymentStatus `json:"payment_status"`
	PaymentMethod          string        `json:"payment_method"`
	PaymentProvider        string        `json:"payment_provider"`
	PaymentReference       string        `json:"payment_reference"`
	PaymentExternalOrderID string        `json:"payment_external_order_id"`
	PaymentQRString        string        `json:"payment_qr_string"`
	PaymentQRURL           string        `json:"payment_qr_url"`
	PaymentExpiryTime      *time.Time    `json:"payment_expiry_time,omitempty"`
	PaymentUpdatedAt       *time.Time    `json:"payment_updated_at,omitempty"`
	Notes                  string        `json:"notes"`
	CreatedAt              time.Time     `json:"created_at"`
	EstimatedDoneAt        time.Time     `json:"estimated_done_at"`
	Items                  []OrderItem   `json:"items"`
	Media                  []Media       `json:"media"`
	Timeline               []Timeline    `json:"timeline"`
}

type Timeline struct {
	ID      int       `json:"-"`
	OrderID int       `json:"-"`
	Label   string    `json:"label"`
	Done    bool      `json:"done"`
	Time    time.Time `json:"time,omitempty"`
}

type Dashboard struct {
	TotalToday       int            `json:"total_today"`
	ActiveOrders     int            `json:"active_orders"`
	CompletedOrders  int            `json:"completed_orders"`
	WaitingPickup    int            `json:"waiting_pickup"`
	DailyRevenue     int            `json:"daily_revenue"`
	MonthlyRevenue   int            `json:"monthly_revenue"`
	StatusCounts     map[Status]int `json:"status_counts"`
	ServiceAnalytics map[string]int `json:"service_analytics"`
	RecentOrders     []Order        `json:"recent_orders"`
	Activities       []string       `json:"activities"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateOrderRequest struct {
	CustomerName    string        `json:"customer_name"`
	CustomerPhone   string        `json:"customer_phone"`
	Status          Status        `json:"status"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	PaymentMethod   string        `json:"payment_method"`
	Notes           string        `json:"notes"`
	EstimatedDoneAt time.Time     `json:"estimated_done_at"`
	Items           []OrderItem   `json:"items"`
}

type PaymentUpdate struct {
	PaymentStatus          PaymentStatus
	PaymentMethod          string
	PaymentProvider        string
	PaymentReference       string
	PaymentExternalOrderID string
	PaymentQRString        string
	PaymentQRURL           string
	PaymentExpiryTime      *time.Time
	PaymentUpdatedAt       *time.Time
}
