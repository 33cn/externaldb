package main

import "net/http"

// handleHealth 健康检查
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "OK",
		Data:    map[string]string{"status": "healthy"},
	})
}

