package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	httpapi "github.com/AYaSmyslov/faqapi/internal/http"
	"github.com/AYaSmyslov/faqapi/internal/models"
	"github.com/AYaSmyslov/faqapi/internal/service"
	"github.com/stretchr/testify/require"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type createdQuestion struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func setupTestServer(t *testing.T) http.Handler {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Question{}, &models.Answer{}))

	svc := service.NewFAQService(db)
	srv := httpapi.NewServer(svc)

	return srv
}

func TestCreateAndGetQuestion(t *testing.T) {
	handler := setupTestServer(t)

	body := map[string]string{"text": "Test question"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/questions/", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var created createdQuestion
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	req2 := httptest.NewRequest(http.MethodGet, "/questions/"+strconv.Itoa(created.ID), nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	require.Equal(t, http.StatusOK, rec2.Code)
}
