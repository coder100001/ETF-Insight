package docs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterSwaggerRoutes 注册 Swagger 相关路由
func RegisterSwaggerRoutes(router *gin.Engine) {
	// Swagger JSON
	router.GET("/swagger.json", SwaggerJSONHandler)

	// Swagger UI
	router.GET("/swagger", SwaggerUIHandler)
	router.GET("/swagger/*any", SwaggerUIHandler)
}

// SwaggerJSONHandler 返回 Swagger JSON 规范
func SwaggerJSONHandler(c *gin.Context) {
	spec := SwaggerSpec()
	c.JSON(http.StatusOK, spec)
}

// SwaggerUIHandler 返回 Swagger UI HTML
func SwaggerUIHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerUIHTML)
}

// swaggerUIHTML Swagger UI HTML 模板
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ETF-Insight API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.9.0/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.9.0/favicon-16x16.png" sizes="16x16" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin: 0;
            background: #fafafa;
        }
        .topbar {
            background: #1a1a1a !important;
        }
        .topbar .download-url-wrapper input[type=text] {
            background: #333;
            color: #fff;
        }
        .topbar .download-url-wrapper .download-url-button {
            background: #4990e2;
        }
        .information-container .info .title {
            color: #1a1a1a;
        }
        .scheme-container {
            background: #fff;
            margin: 20px 0;
            padding: 20px;
            border-radius: 4px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log('Swagger UI loaded successfully');
                },
                onFailure: function(data) {
                    console.error('Unable to Load SwaggerUI', data);
                }
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`
