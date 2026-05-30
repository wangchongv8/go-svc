// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package health

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"go-svc/apps/gateway-api/internal/handler/httperr"
	"go-svc/apps/gateway-api/internal/logic/health"
	"go-svc/apps/gateway-api/internal/svc"
)

func HealthzHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := health.NewHealthzLogic(r.Context(), svcCtx)
		err := l.Healthz()
		if err != nil {
			httperr.Write(w, err)
		} else {
			httpx.OkJson(w, map[string]string{"status": "ok"})
		}
	}
}
