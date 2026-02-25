# ════════════════════════════════════════════════════════════════
#  Server Makefile — upnixo.com
#  Root: server/
#  Dev  : make nginx-dev
#  Prod : make nginx-deploy
# ════════════════════════════════════════════════════════════════

NGINX_CONF_DIR := /etc/nginx/conf.d
NGINX_SSL_DIR  := /etc/nginx/ssl
APP_CONF       := api.upnixo.conf

NGINX_DIR      := ./nginx
SSL_SRC        := $(NGINX_DIR)/ssl
PROD_CONF      := $(NGINX_DIR)/prod/nginx.conf
DEV_CONF       := $(NGINX_DIR)/dev/nginx.conf

.PHONY: help nginx-dev nginx-deploy ssl-setup ssl-check test reload restart logs logs-error status clean

# ── Default: help ─────────────────────────────────────────────
help:
	@echo ""
	@echo "  ╔══════════════════════════════════════════════╗"
	@echo "  ║      Server Makefile — api.upnixo.com        ║"
	@echo "  ╠══════════════════════════════════════════════╣"
	@echo "  ║  make nginx-dev      → Dev config (HTTP)     ║"
	@echo "  ║  make ssl-setup      → SSL keys copy karo    ║"
	@echo "  ║  make ssl-check      → SSL verify karo       ║"
	@echo "  ║  make nginx-deploy   → Production deploy     ║"
	@echo "  ║  make test           → Nginx config test     ║"
	@echo "  ║  make reload         → Zero-downtime reload  ║"
	@echo "  ║  make restart        → Hard restart          ║"
	@echo "  ║  make logs           → Live access logs      ║"
	@echo "  ║  make logs-error     → Live error logs       ║"
	@echo "  ║  make status         → Nginx status          ║"
	@echo "  ║  make clean          → Config hata do        ║"
	@echo "  ╚══════════════════════════════════════════════╝"
	@echo ""

# ════════════════════════════════════════════════════════════════
#  DEV — Simple HTTP
# ════════════════════════════════════════════════════════════════
nginx-dev:
	@echo "🔧 Dev config laga raha hoon..."
	@sudo cp $(DEV_CONF) $(NGINX_CONF_DIR)/$(APP_CONF)
	@$(MAKE) test
	@$(MAKE) reload
	@echo ""
	@echo "✅ Dev ready!"
	@echo "   http://localhost/api/v1/auth"
	@echo "   http://localhost/api/v1/user"
	@echo "   http://localhost/api/v1/billing"
	@echo "   http://localhost/api/v1/webhook"
	@echo "   http://localhost/health"

# ════════════════════════════════════════════════════════════════
#  SSL SETUP — Cloudflare Origin Certificate
# ════════════════════════════════════════════════════════════════
ssl-setup:
	@echo "🔐 SSL setup shuru..."

	@# Check karo files exist hain
	@if [ ! -f $(SSL_SRC)/api.upnixo.com.crt ]; then \
		echo "❌ nginx/ssl/api.upnixo.com.crt nahi mila!"; \
		exit 1; \
	fi
	@if [ ! -f $(SSL_SRC)/api.upnixo.com.key ]; then \
		echo "❌ nginx/ssl/api.upnixo.com.key nahi mila!"; \
		exit 1; \
	fi

	@# SSL directory banao
	@sudo mkdir -p $(NGINX_SSL_DIR)

	@# Keys copy karo
	@sudo cp $(SSL_SRC)/api.upnixo.com.crt $(NGINX_SSL_DIR)/api.upnixo.com.crt
	@sudo cp $(SSL_SRC)/api.upnixo.com.key $(NGINX_SSL_DIR)/api.upnixo.com.key

	@# Permissions (IMPORTANT!)
	@sudo chmod 600 $(NGINX_SSL_DIR)/api.upnixo.com.key
	@sudo chmod 644 $(NGINX_SSL_DIR)/api.upnixo.com.crt
	@sudo chown root:root $(NGINX_SSL_DIR)/api.upnixo.com.*

	@echo "✅ SSL keys /etc/nginx/ssl/ mein copy ho gaye!"
	@$(MAKE) ssl-check

ssl-check:
	@echo "🔍 SSL verify kar raha hoon..."
	@if sudo test -f $(NGINX_SSL_DIR)/api.upnixo.com.crt; then \
		echo "✅ Certificate: $(NGINX_SSL_DIR)/api.upnixo.com.crt"; \
		sudo openssl x509 -in $(NGINX_SSL_DIR)/api.upnixo.com.crt -noout -subject -dates 2>/dev/null || true; \
	else \
		echo "❌ Certificate nahi mila — pehle: make ssl-setup"; \
	fi
	@if sudo test -f $(NGINX_SSL_DIR)/api.upnixo.com.key; then \
		echo "✅ Private key: $(NGINX_SSL_DIR)/api.upnixo.com.key"; \
	else \
		echo "❌ Private key nahi mila — pehle: make ssl-setup"; \
	fi

# ════════════════════════════════════════════════════════════════
#  PRODUCTION DEPLOY
# ════════════════════════════════════════════════════════════════
nginx-deploy:
	@echo "🚀 Production deploy shuru..."

	@# SSL check
	@if ! sudo test -f $(NGINX_SSL_DIR)/api.upnixo.com.crt || \
	    ! sudo test -f $(NGINX_SSL_DIR)/api.upnixo.com.key; then \
		echo "❌ SSL keys nahi hain! Pehle chalao: make ssl-setup"; \
		exit 1; \
	fi

	@sudo cp $(PROD_CONF) $(NGINX_CONF_DIR)/$(APP_CONF)
	@echo "✅ Production config copy hua"
	@$(MAKE) test
	@$(MAKE) reload
	@echo ""
	@echo "🎉 Production live!"
	@echo "   https://api.upnixo.com/api/v1/auth"
	@echo "   https://api.upnixo.com/api/v1/user"
	@echo "   https://api.upnixo.com/health"

# ════════════════════════════════════════════════════════════════
#  HELPERS
# ════════════════════════════════════════════════════════════════
test:
	@echo "🧪 Nginx config test..."
	@sudo nginx -t
	@echo "✅ Config OK"

reload:
	@echo "🔄 Nginx reload (zero downtime)..."
	@sudo systemctl reload nginx

restart:
	@sudo systemctl restart nginx
	@sudo systemctl status nginx --no-pager

logs:
	@sudo tail -f /var/log/nginx/api_upnixo_access.log

logs-error:
	@sudo tail -f /var/log/nginx/api_upnixo_error.log

status:
	@sudo systemctl status nginx --no-pager

clean:
	@sudo rm -f $(NGINX_CONF_DIR)/$(APP_CONF)
	@echo "✅ Config removed. Reload karo: make reload"