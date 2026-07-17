# Locales moved to the backend (GET /api/locales); this service now serves only
# generated avatars and the catalog media volume (mounted at runtime).
FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY avatars /usr/share/nginx/html/avatars
EXPOSE 80
