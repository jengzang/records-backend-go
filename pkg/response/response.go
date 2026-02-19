package response

import "github.com/gin-gonic/gin"

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error sends an error response
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest sends a 400 bad request response
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

// NotFound sends a 404 not found response
func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

// InternalError sends a 500 internal server error response
func InternalError(c *gin.Context, message string) {
	Error(c, 500, message)
}
