package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ad(t *testing.T) {
	//https://juejin.cn/post/7140302505006596133
	//使用gin的Router
	r := setupRoute()
	// 準備測試資料
	data := url.Values{"title": {"AD test"},
		"startAt":  {"2023-12-10T03:00:00.000Z"},
		"endAt":    {"2023-12-31T16:00:00.000Z"},
		"ageStart": {"20"},
		"ageEnd":   {"30"},
		"country":  {"TW", "JP"},
		"platform": {"android", "ios"},
	}

	reqbody := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, "/api/v1/ad", reqbody)
	if err != nil {
		t.Fatalf("Request failed, err: %v", err)
	}
	// 設定request Header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 紀錄回傳內容
	rec := httptest.NewRecorder()
	// 使用http服務
	r.ServeHTTP(rec, req)
	assert.EqualValues(t, 200, rec.Code)
}
func Test_public(t *testing.T) {
	//使用gin的Router
	r := setupRoute()
	// 準備測試資料
	req, err := http.NewRequest("GET", "/api/v1/ad/get", nil)
	if err != nil {
		t.Fatalf("Request failed, err: %v", err)
	}
	q := req.URL.Query()
	q.Add("offset", "0")
	q.Add("limit", "10")
	q.Add("age", "1000")
	q.Add("gender", "M")
	q.Add("coountry", "TW")
	q.Add("platform", "ios")
	req.URL.RawQuery = q.Encode()
	// 紀錄回傳內容
	w := httptest.NewRecorder()
	// 使用http服務
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func Test_public_fail(t *testing.T) {
	//使用gin的Router
	r := setupRoute()
	// 準備測試資料
	req, err := http.NewRequest("GET", "/api/v1/ad/get", nil)
	if err != nil {
		t.Fatalf("Request failed, err: %v", err)
	}
	q := req.URL.Query()
	q.Add("offset", "0")
	//設置負數
	q.Add("limit", "-1000")
	q.Add("age", "1000")
	q.Add("gender", "M")
	q.Add("coountry", "TW")
	q.Add("platform", "ios")
	req.URL.RawQuery = q.Encode()
	// 紀錄回傳內容
	w := httptest.NewRecorder()
	// 使用http服務
	r.ServeHTTP(w, req)

	assert.Equal(t, 202, w.Code)
}
