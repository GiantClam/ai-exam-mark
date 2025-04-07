package utils

import (
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"io"
	"os"
)

// Min 返回两个整数中较小的那个
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max 返回两个整数中较大的那个
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RespondWithError 用于统一返回错误信息
func RespondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
		"is_error": true,
	})
}

// SaveUploadedFile 保存上传的文件到指定路径
func SaveUploadedFile(fileHeader *multipart.FileHeader, dst string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// 复制文件内容
	_, err = io.Copy(out, src)
	return err
}
