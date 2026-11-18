package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders menambahkan HTTP security headers ke setiap response.
// Header ini melindungi dari serangan umum: clickjacking, MIME sniffing, XSS, dll.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()

		// Cegah MIME sniffing — browser harus mempercayai Content-Type yang dikirim server.
		h.Set("X-Content-Type-Options", "nosniff")

		// Cegah clickjacking — halaman ini tidak boleh di-iframe.
		h.Set("X-Frame-Options", "DENY")

		// Aktifkan XSS protection di browser lama (modern browser sudah built-in).
		h.Set("X-XSS-Protection", "1; mode=block")

		// Batasi referrer yang dikirim saat navigasi cross-origin.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS — paksa HTTPS selama 1 tahun (hanya efektif via HTTPS).
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content-Security-Policy — batasi sumber script/style yang boleh dimuat.
		h.Set("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}
