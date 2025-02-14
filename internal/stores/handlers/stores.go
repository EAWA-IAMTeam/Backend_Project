package handlers

import (
	"backend_project/internal/stores/services"
	"net/http"

	"github.com/labstack/echo"
)

// type StoreHandler interface {
// 	LazadaLinkStore(ctx echo.Context) error
// }

type StoreHandler struct {
	// *Handler [logger]
	StoreService *services.StoreService
}

func NewStoreHandler(ss *services.StoreService) *StoreHandler {
	return &StoreHandler{StoreService: ss}
}

// LazadaGenerateAccessToken handles fetching and storing Lazada access tokens
func (sh *StoreHandler) LazadaLinkStore(ctx echo.Context) error {
	authCode := ctx.QueryParam("code")
	if authCode == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Authorization code is required"})
	}

	response, err := sh.StoreService.FetchStoreInfo(authCode)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return ctx.JSON(http.StatusOK, response)
}
