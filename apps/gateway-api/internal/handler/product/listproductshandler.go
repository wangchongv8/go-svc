// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package product

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go-svc/apps/gateway-api/internal/handler/httperr"
	"go-svc/apps/gateway-api/internal/logic/product"
	"go-svc/apps/gateway-api/internal/svc"
)

func ListProductsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := product.NewListProductsLogic(r.Context(), svcCtx)
		resp, err := l.ListProducts()
		if err != nil {
			httperr.Write(w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
