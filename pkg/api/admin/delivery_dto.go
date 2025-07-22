package admin

// CreateTransactionalDeliveryRequest is the request for creating a transactional delivery.
type CreateTransactionalDeliveryRequest struct {
	Name    string                 `json:"name"`
	Email   string                 `json:"email"`
	Subject string                 `json:"subject"`
	Data    map[string]interface{} `json:"data"`
	Headers map[string]string      `json:"headers"`
	Tags    []string               `json:"tags"`
}
