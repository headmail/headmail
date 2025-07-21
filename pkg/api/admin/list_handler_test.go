package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockListService is a mock implementation of ListService.
type MockListService struct {
	mock.Mock
}

func (m *MockListService) CreateList(ctx context.Context, list *domain.List) error {
	args := m.Called(ctx, list)
	// Simulate service layer logic: copy generated ID back to the input object
	list.ID = "mock-id"
	return args.Error(0)
}

func TestListHandler_createList(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)

	router := chi.NewRouter()
	router.Route("/", handler.RegisterRoutes)

	newList := domain.List{
		Name:        "Test Newsletter",
		Description: "For testing",
		Tags:        []string{"test"},
	}
	body, _ := json.Marshal(newList)

	// Setup mock expectation
	mockService.On("CreateList", mock.Anything, mock.AnythingOfType("*domain.List")).Return(nil)

	req, err := http.NewRequest("POST", "/lists", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdList domain.List
	err = json.Unmarshal(rr.Body.Bytes(), &createdList)
	require.NoError(t, err)

	assert.Equal(t, "mock-id", createdList.ID)
	assert.Equal(t, "Test Newsletter", createdList.Name)

	mockService.AssertExpectations(t)
}
