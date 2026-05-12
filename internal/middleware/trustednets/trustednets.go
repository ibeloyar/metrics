package trustednets

import (
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
)

func TrustedSubnetMiddleware(trustedSubnet string, lg *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				lg.Warn("missing X-Real-IP header")
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			ok, err := ipInCIDR(realIP, trustedSubnet)
			if err != nil {
				lg.Warnw("invalid trusted subnet or IP", "error", err, "X-Real-IP", realIP, "trusted_subnet", trustedSubnet)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			if !ok {
				lg.Infow("IP not in trusted subnet", "X-Real-IP", realIP, "trusted_subnet", trustedSubnet)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ipInCIDR(ipStr, cidrStr string) (bool, error) {
	if cidrStr == "" {
		return true, nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	_, ipnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, err
	}

	return ipnet.Contains(ip), nil
}
