package web

import (
	"net/http"
	"speedtest/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

//var sessionKey = "session_key" // for debug only (change it to a random string in production)

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		//if session.Get("authenticated") != true && c.Request.URL.Path != "/api/login" && c.Request.URL.Path != "/login.html" && c.Request.URL.Path != "/login" && !strings.HasPrefix(c.Request.URL.Path, "/backend") {
		if session.Get("authenticated") != true && c.Request.URL.Path != "/api/login" && c.Request.URL.Path != "/login.html" && c.Request.URL.Path != "/login" {
			c.Redirect(http.StatusFound, "/login.html")
			c.Abort()
			return
		} /*else if session.Get("authenticated") == true {
			// 记录路径
			logInfo("passed path: %s", c.Request.URL.Path)
		} */ //for debug
		c.Next()
	}
}

func AuthLogin(c *gin.Context, cfg *config.Config) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 这里应该验证用户名和密码
	if username == cfg.Auth.Username && password == cfg.Auth.Password {
		session := sessions.Default(c)
		session.Set("authenticated", true)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"success": true})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的凭证"})
	}
}

func AuthLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
