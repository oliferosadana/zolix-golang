package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	store     DataStore
	staticDir string
	assetsDir string
	uploadDir string
	jwtSecret string
	publicURL string
	wa        *WhatsAppClient
	payments  *AutoGoPayClient
}

type DataStore interface {
	Dashboard() Dashboard
	Orders(query, status string) []Order
	Order(id int) (Order, bool)
	OrderByInvoice(invoice string) (Order, bool)
	CreateOrder(input Order) (Order, error)
	Customers() []Customer
	Services() []Service
	Users() []User
	Authenticate(email, password string) (User, bool)
	UpdateOrder(id int, input UpdateOrderRequest) (Order, error)
	DeleteOrder(id int) bool
	AddMedia(orderID int, mediaType, url string) (Media, error)
	UpdateOrderPayment(id int, update PaymentUpdate) (Order, error)
	OrderByPaymentReference(reference string) (Order, bool)
}

func NewServer(store DataStore, staticDir, assetsDir string) *Server {
	return &Server{
		store:     store,
		staticDir: staticDir,
		assetsDir: assetsDir,
		uploadDir: envDefault("UPLOAD_DIR", "uploads"),
		jwtSecret: envDefault("JWT_SECRET", "zolix-dev-secret"),
		publicURL: envDefault("PUBLIC_BASE_URL", "http://localhost:8080"),
		wa:        NewWhatsAppClient(),
		payments:  NewAutoGoPayClient(),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/order/", s.trackingPage)
	mux.HandleFunc("/", s.index)
	mux.HandleFunc("/api/v1/auth/login", s.login)
	mux.HandleFunc("/api/v1/public/orders/", s.publicOrderByInvoice)
	mux.HandleFunc("/api/v1/public/payment-config", s.publicPaymentConfig)
	mux.HandleFunc("/api/v1/public/payments/qris/generate", s.publicGenerateQRISPayment)
	mux.HandleFunc("/api/v1/public/payments/qris/status", s.publicCheckQRISPayment)
	mux.HandleFunc("/api/v1/public/payments/method", s.publicSetPaymentMethod)
	mux.HandleFunc("/api/v1/auth/me", s.requireAuth(s.me))
	mux.HandleFunc("/api/v1/dashboard", s.requireAuth(s.dashboard))
	mux.HandleFunc("/api/v1/orders", s.requireAuth(s.orders))
	mux.HandleFunc("/api/v1/orders/", s.requireAuth(s.orderByID))
	mux.HandleFunc("/api/v1/customers", s.requireAuth(s.customers))
	mux.HandleFunc("/api/v1/services", s.requireAuth(s.services))
	mux.HandleFunc("/api/v1/upload", s.requireAuth(s.upload))
	mux.HandleFunc("/api/v1/whatsapp/send", s.requireAuth(s.sendWhatsApp))
	mux.HandleFunc("/api/v1/payments/qris/generate", s.requireAuth(s.generateQRISPayment))
	mux.HandleFunc("/api/v1/payments/qris/status", s.requireAuth(s.checkQRISPayment))
	mux.HandleFunc("/api/v1/payments/autogopay/webhook", s.autoGoPayWebhook)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(s.assetsDir))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.uploadDir))))
	return mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		notFound(w)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.staticDir, "index.html"))
}

func (s *Server) trackingPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.staticDir, "tracking.html"))
}

func (s *Server) publicOrderByInvoice(w http.ResponseWriter, r *http.Request) {
	invoice := strings.TrimPrefix(r.URL.Path, "/api/v1/public/orders/")
	if invoice == "" {
		writeError(w, http.StatusBadRequest, "invoice is required")
		return
	}
	order, ok := s.store.OrderByInvoice(invoice)
	if !ok {
		notFound(w)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) publicPaymentConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"qris_enabled": s.payments.Enabled(),
		"transfer": map[string]string{
			"bank_name":      envDefault("TRANSFER_BANK_NAME", "BCA"),
			"account_number": envDefault("TRANSFER_ACCOUNT_NUMBER", "Isi nomor rekening di .env"),
			"account_name":   envDefault("TRANSFER_ACCOUNT_NAME", "Zolix Shoe Care"),
			"instructions":   envDefault("TRANSFER_INSTRUCTIONS", "Transfer sesuai total invoice lalu kirim bukti pembayaran melalui WhatsApp."),
		},
		"cash": map[string]string{
			"instructions": envDefault("CASH_INSTRUCTIONS", "Bayar langsung di outlet saat pickup atau saat menyerahkan sepatu."),
		},
	})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	user, ok := s.store.Authenticate(input.Email, input.Password)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	token, err := s.signToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Dashboard())
}

func (s *Server) orders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.Orders(r.URL.Query().Get("q"), r.URL.Query().Get("status")))
	case http.MethodPost:
		var input Order
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		order, err := s.store.CreateOrder(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, order)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) orderByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/v1/orders/"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order ID")
		return
	}
	switch r.Method {
	case http.MethodGet:
		order, ok := s.store.Order(id)
		if !ok {
			notFound(w)
			return
		}
		writeJSON(w, http.StatusOK, order)
	case http.MethodPut:
		var input UpdateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		order, err := s.store.UpdateOrder(id, input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, order)
	case http.MethodDelete:
		if !s.store.DeleteOrder(id) {
			notFound(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) customers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Customers())
}

func (s *Server) services(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Services())
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart payload")
		return
	}
	orderID, err := strconv.Atoi(r.FormValue("order_id"))
	if err != nil || orderID <= 0 {
		writeError(w, http.StatusBadRequest, "valid order_id is required")
		return
	}
	mediaType := strings.TrimSpace(r.FormValue("type"))
	if mediaType != "before" && mediaType != "after" {
		writeError(w, http.StatusBadRequest, "type must be before or after")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		writeError(w, http.StatusBadRequest, "supported formats: JPG, PNG, WEBP")
		return
	}
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create upload directory")
		return
	}
	name := fmt.Sprintf("order-%d-%s-%d%s", orderID, mediaType, time.Now().UnixNano(), ext)
	targetPath := filepath.Join(s.uploadDir, name)
	target, err := os.Create(targetPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save upload")
		return
	}
	defer target.Close()
	written, err := io.Copy(target, io.LimitReader(file, 10<<20+1))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to write upload")
		return
	}
	if written > 10<<20 {
		_ = os.Remove(targetPath)
		writeError(w, http.StatusBadRequest, "file is too large")
		return
	}
	media, err := s.store.AddMedia(orderID, mediaType, "/uploads/"+name)
	if err != nil {
		_ = os.Remove(targetPath)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, media)
}

func (s *Server) generateQRISPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		OrderID int `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	order, ok := s.store.Order(input.OrderID)
	if !ok {
		notFound(w)
		return
	}
	updated, err := s.generateQRISForOrder(r.Context(), order)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"order":   updated,
		"payment": paymentPayload(updated),
	})
}

func (s *Server) publicGenerateQRISPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		InvoiceNumber string `json:"invoice_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	order, ok := s.store.OrderByInvoice(input.InvoiceNumber)
	if !ok {
		notFound(w)
		return
	}
	updated, err := s.generateQRISForOrder(r.Context(), order)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"order":   updated,
		"payment": paymentPayload(updated),
	})
}

func (s *Server) checkQRISPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		OrderID int `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	order, ok := s.store.Order(input.OrderID)
	if !ok {
		notFound(w)
		return
	}
	updated, status, err := s.checkQRISForOrder(r.Context(), order)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"order":   updated,
		"payment": paymentPayload(updated),
		"status":  status,
	})
}

func (s *Server) publicCheckQRISPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		InvoiceNumber string `json:"invoice_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	order, ok := s.store.OrderByInvoice(input.InvoiceNumber)
	if !ok {
		notFound(w)
		return
	}
	updated, status, err := s.checkQRISForOrder(r.Context(), order)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"order":   updated,
		"payment": paymentPayload(updated),
		"status":  status,
	})
}

func (s *Server) publicSetPaymentMethod(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		InvoiceNumber string `json:"invoice_number"`
		Method        string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	order, ok := s.store.OrderByInvoice(input.InvoiceNumber)
	if !ok {
		notFound(w)
		return
	}
	method := strings.TrimSpace(input.Method)
	if method != "Transfer" && method != "Cash" {
		writeError(w, http.StatusBadRequest, "method must be Transfer or Cash")
		return
	}
	now := time.Now()
	updated, err := s.store.UpdateOrderPayment(order.ID, PaymentUpdate{
		PaymentStatus:    PaymentUnpaid,
		PaymentMethod:    method,
		PaymentProvider:  "Manual",
		PaymentUpdatedAt: &now,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"order":   updated,
		"payment": paymentPayload(updated),
	})
}

func (s *Server) autoGoPayWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook body")
		return
	}
	if !s.payments.VerifySignature(body, r.Header.Get("X-Signature")) {
		writeError(w, http.StatusUnauthorized, "invalid webhook signature")
		return
	}
	var payload AutoGoPayWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook JSON")
		return
	}
	if strings.ToLower(strings.TrimSpace(payload.Event)) == "verification.challenge" {
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
		return
	}
	switch strings.ToLower(strings.TrimSpace(payload.Event)) {
	case "payment.update", "transaction.received":
	default:
		writeJSON(w, http.StatusOK, map[string]any{"ignored": true})
		return
	}
	order, ok := s.store.OrderByPaymentReference(payload.Transaction.ID)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "matched": false})
		return
	}
	updated, err := s.applyAutoGoPayStatus(order, payload.Transaction.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"received": true,
		"matched":  true,
		"order":    updated.ID,
	})
}

func (s *Server) applyAutoGoPayStatus(order Order, status string) (Order, error) {
	now := time.Now()
	update := PaymentUpdate{
		PaymentStatus:          order.PaymentStatus,
		PaymentMethod:          "QRIS",
		PaymentProvider:        "AutoGoPay",
		PaymentReference:       order.PaymentReference,
		PaymentExternalOrderID: order.PaymentExternalOrderID,
		PaymentQRString:        order.PaymentQRString,
		PaymentQRURL:           order.PaymentQRURL,
		PaymentExpiryTime:      order.PaymentExpiryTime,
		PaymentUpdatedAt:       &now,
	}
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "paid", "settlement", "success", "completed":
		update.PaymentStatus = PaymentPaid
	case "expired", "expire", "failed", "cancelled", "canceled", "deny":
		update.PaymentStatus = PaymentUnpaid
	}
	return s.store.UpdateOrderPayment(order.ID, update)
}

func (s *Server) generateQRISForOrder(ctx context.Context, order Order) (Order, error) {
	if order.TotalPrice <= 0 {
		return Order{}, errors.New("order total must be greater than zero")
	}
	if !s.payments.Enabled() {
		return Order{}, errors.New("AUTOGOPAY_API_KEY is not configured")
	}
	result, err := s.payments.GenerateQRIS(ctx, order.TotalPrice)
	if err != nil {
		return Order{}, err
	}
	now := time.Now()
	return s.store.UpdateOrderPayment(order.ID, PaymentUpdate{
		PaymentStatus:          PaymentUnpaid,
		PaymentMethod:          "QRIS",
		PaymentProvider:        "AutoGoPay",
		PaymentReference:       result.TransactionID,
		PaymentExternalOrderID: result.OrderID,
		PaymentQRString:        result.QRString,
		PaymentQRURL:           result.QRURL,
		PaymentExpiryTime:      result.ExpiryTime,
		PaymentUpdatedAt:       &now,
	})
}

func (s *Server) checkQRISForOrder(ctx context.Context, order Order) (Order, string, error) {
	if strings.TrimSpace(order.PaymentReference) == "" {
		return Order{}, "", errors.New("order does not have an AutoGoPay transaction")
	}
	if !s.payments.Enabled() {
		return Order{}, "", errors.New("AUTOGOPAY_API_KEY is not configured")
	}
	status, err := s.payments.QRISStatus(ctx, order.PaymentReference)
	if err != nil {
		return Order{}, "", err
	}
	updated, err := s.applyAutoGoPayStatus(order, status.Status)
	if err != nil {
		return Order{}, "", err
	}
	return updated, status.Status, nil
}

func paymentPayload(order Order) map[string]any {
	return map[string]any{
		"provider":          order.PaymentProvider,
		"transaction_id":    order.PaymentReference,
		"external_order_id": order.PaymentExternalOrderID,
		"qr_string":         order.PaymentQRString,
		"qr_url":            order.PaymentQRURL,
		"expiry_time":       order.PaymentExpiryTime,
		"status":            order.PaymentStatus,
		"updated_at":        order.PaymentUpdatedAt,
	}
}

func (s *Server) sendWhatsApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input WhatsAppSendRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	order, ok := s.store.Order(input.OrderID)
	if !ok {
		notFound(w)
		return
	}
	message := strings.TrimSpace(input.Message)
	if message == "" {
		message = orderWhatsAppMessage(order, s.publicURL)
	}
	if err := s.wa.SendText(order.CustomerPhone, message); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sent": true,
		"to":   whatsappChatID(order.CustomerPhone),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func notFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, "resource not found")
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

type contextKey string

const userContextKey contextKey = "user"

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		user, err := s.verifyToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	}
}

func userFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

func (s *Server) signToken(user User) (string, error) {
	expiry := time.Now().Add(24 * time.Hour).Unix()
	payload := fmt.Sprintf("%d|%s|%d", user.ID, user.Email, expiry)
	mac := hmac.New(sha256.New, []byte(s.jwtSecret))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", err
	}
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signature, nil
}

func (s *Server) verifyToken(token string) (User, error) {
	if token == "" {
		return User{}, errors.New("missing token")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return User{}, errors.New("invalid token")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return User{}, err
	}
	payload := string(payloadBytes)
	mac := hmac.New(sha256.New, []byte(s.jwtSecret))
	_, _ = mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return User{}, errors.New("invalid signature")
	}
	fields := strings.Split(payload, "|")
	if len(fields) != 3 {
		return User{}, errors.New("invalid payload")
	}
	expiry, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || time.Now().Unix() > expiry {
		return User{}, errors.New("expired token")
	}
	for _, user := range s.store.Users() {
		if user.Email == fields[1] {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func envDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
