package services

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Min 返回两个整数中较小的那个
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CleanMarkdownCodeBlock 从Markdown代码块中提取纯JSON内容
func CleanMarkdownCodeBlock(content string) string {
	// 记录原始内容的长度以进行调试
	originalLength := len(content)
	log.Printf("[DEBUG] 开始清理Markdown代码块")
	log.Printf("[DEBUG] 原始内容长度: %d 字符", originalLength)

	// 如果内容为空，直接返回
	if content == "" {
		log.Printf("[WARN] 内容为空")
		return "{}"
	}

	// 打印前100个字符，用于调试
	previewLength := Min(100, originalLength)
	log.Printf("[DEBUG] 原始内容前%d个字符: %s", previewLength, content[:previewLength])

	// 先检查是否包含反引号，这可能表示是代码块
	if strings.Contains(content, "```") {
		log.Printf("[DEBUG] 检测到代码块标记")

		// 更完善的正则表达式，匹配可能出现的各种代码块格式
		// 移除开始的代码块标记，包括可能的空格和换行
		startPattern := `(?s)^[\s\n]*` + "```" + `(?:json|javascript|js)?[\s\n]*`
		content = regexp.MustCompile(startPattern).ReplaceAllString(content, "")
		log.Printf("[DEBUG] 移除了开始的代码块标记")

		// 移除结束的代码块标记，包括可能的空格和换行
		endPattern := `(?s)[\s\n]*` + "```" + `[\s\n]*$`
		content = regexp.MustCompile(endPattern).ReplaceAllString(content, "")
		log.Printf("[DEBUG] 移除了结束的代码块标记")
	}

	// 处理单行代码（使用单个反引号的情况）
	if strings.Contains(content, "`") {
		log.Printf("[DEBUG] 检测到单行代码标记")
		content = regexp.MustCompile("`").ReplaceAllString(content, "")
		log.Printf("[DEBUG] 移除了单行代码标记")
	}

	// 清理内容
	content = strings.TrimSpace(content)
	log.Printf("[DEBUG] 清理了空白字符")

	// 确保内容以JSON对象开始
	if !strings.HasPrefix(content, "{") {
		log.Printf("[WARN] 内容不以 { 开始，尝试查找第一个 {")
		jsonStart := strings.Index(content, "{")
		if jsonStart >= 0 {
			content = content[jsonStart:]
			log.Printf("[DEBUG] 已找到第一个 {，移除前面的内容")
		} else {
			log.Printf("[WARN] 未找到 {，返回空对象")
			return "{}"
		}
	}

	// 确保内容以JSON对象结束
	if !strings.HasSuffix(content, "}") {
		log.Printf("[WARN] 内容不以 } 结束，尝试查找最后一个 }")
		jsonEnd := strings.LastIndex(content, "}")
		if jsonEnd >= 0 && jsonEnd < len(content)-1 {
			content = content[:jsonEnd+1]
			log.Printf("[DEBUG] 已找到最后一个 }，移除后面的内容")
		} else {
			log.Printf("[WARN] 未找到 }，返回空对象")
			return "{}"
		}
	}

	// 验证JSON格式
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(content), &jsonObj); err != nil {
		log.Printf("[ERROR] JSON格式无效: %v", err)
		return "{}"
	}

	// 打印清理后的内容长度以进行调试
	cleanedLength := len(content)
	log.Printf("[DEBUG] 清理后内容长度: %d (减少了 %d 个字符)", cleanedLength, originalLength-cleanedLength)

	return content
}

// SanitizeUTF8 清理字符串中的无效UTF-8字符
func SanitizeUTF8(s string) string {
	if s == "" {
		return s
	}

	// 先检查字符串是否有效的UTF-8
	if !utf8.ValidString(s) {
		var builder strings.Builder
		builder.Grow(len(s))

		// 遍历字符串，只保留有效的UTF-8字符
		for i := 0; i < len(s); {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r != utf8.RuneError || size == 1 {
				builder.WriteRune(r)
			}
			i += size
		}
		return builder.String()
	}
	return s
}

// TryFixJsonFormat 尝试修复JSON格式
func TryFixJsonFormat(jsonStr string) string {
	// 移除可能存在的非JSON前缀
	jsonStartIndex := strings.Index(jsonStr, "{")
	if jsonStartIndex > 0 {
		jsonStr = jsonStr[jsonStartIndex:]
	}

	// 确保JSON结尾是正确的
	jsonEndIndex := strings.LastIndex(jsonStr, "}")
	if jsonEndIndex >= 0 && jsonEndIndex < len(jsonStr)-1 {
		jsonStr = jsonStr[:jsonEndIndex+1]
	}

	// 处理开闭括号不匹配的情况
	openCount := strings.Count(jsonStr, "{")
	closeCount := strings.Count(jsonStr, "}")

	if openCount > closeCount {
		// 补充缺少的闭括号
		jsonStr += strings.Repeat("}", openCount-closeCount)
	}

	return jsonStr
}

// EnsureValidJSON 确保字符串为有效的JSON
// 整合了多种清理和修复技术
func EnsureValidJSON(input string) string {
	// 记录日志，便于调试
	log.Printf("[DEBUG] 开始处理非标准JSON输入")
	log.Printf("[DEBUG] 输入长度: %d 字符", len(input))
	if len(input) > 100 {
		log.Printf("[DEBUG] 输入前100个字符: %s", input[:100])
	} else if len(input) > 0 {
		log.Printf("[DEBUG] 输入内容: %s", input)
	}

	if input == "" {
		log.Printf("[WARN] 输入为空")
		return "{}"
	}

	// 第1步：删除常见的前导说明文本
	// 找到第一个有效的JSON字符，比如 { 或 [
	jsonStart := -1
	for i, char := range input {
		if char == '{' || char == '[' {
			jsonStart = i
			break
		}
	}

	if jsonStart > 0 {
		log.Printf("[INFO] 删除前导文本，位置：%d", jsonStart)
		if jsonStart < 50 {
			log.Printf("[DEBUG] 删除的前导文本: %s", input[:jsonStart])
		}
		input = input[jsonStart:]
	}

	// 第2步：删除后置文本（如果有）
	// 找到最后一个有效的JSON结束字符 } 或 ]
	var searchChar byte
	if strings.HasPrefix(input, "{") {
		searchChar = '}'
	} else if strings.HasPrefix(input, "[") {
		searchChar = ']'
	} else {
		log.Printf("[WARN] 输入不以 { 或 [ 开头，可能不是JSON")
		return "{}"
	}

	// 从后向前搜索结束字符
	jsonEnd := strings.LastIndexByte(input, searchChar)
	if jsonEnd >= 0 && jsonEnd < len(input)-1 {
		log.Printf("[INFO] 删除后置文本，位置：%d", jsonEnd)
		input = input[:jsonEnd+1]
	}

	// 第3步：处理大模型常见的格式问题
	// 3.1 清理可能的Markdown代码块标记
	cleaned := CleanMarkdownCodeBlock(input)

	// 3.2 清理无效的UTF-8字符
	sanitized := SanitizeUTF8(cleaned)

	// 3.3 处理括号不匹配的情况
	openCount := strings.Count(sanitized, "{")
	closeCount := strings.Count(sanitized, "}")
	openBracketCount := strings.Count(sanitized, "[")
	closeBracketCount := strings.Count(sanitized, "]")

	// 修复大括号不匹配
	if openCount != closeCount {
		log.Printf("[WARN] 大括号不匹配: { = %d, } = %d", openCount, closeCount)
		if openCount > closeCount {
			// 补充缺少的闭括号
			sanitized += strings.Repeat("}", openCount-closeCount)
			log.Printf("[INFO] 添加 %d 个缺失的 }", openCount-closeCount)
		} else {
			// 情况较复杂，可能需要更多处理
			log.Printf("[WARN] 闭括号多于开括号，情况复杂")
		}
	}

	// 修复中括号不匹配
	if openBracketCount != closeBracketCount {
		log.Printf("[WARN] 中括号不匹配: [ = %d, ] = %d", openBracketCount, closeBracketCount)
		if openBracketCount > closeBracketCount {
			// 补充缺少的闭中括号
			sanitized += strings.Repeat("]", openBracketCount-closeBracketCount)
			log.Printf("[INFO] 添加 %d 个缺失的 ]", openBracketCount-closeBracketCount)
		} else {
			log.Printf("[WARN] 闭中括号多于开中括号，情况复杂")
		}
	}

	// 第4步：修复常见语法错误
	// 4.1 修复尾部多余的逗号
	sanitized = regexp.MustCompile(`,\s*[}\]]`).ReplaceAllStringFunc(sanitized, func(s string) string {
		return strings.Replace(s, ",", "", 1)
	})

	// 4.2 修复未加引号的键名
	// 这个功能需要更复杂的解析，简单起见只处理最常见的情况
	sanitized = regexp.MustCompile(`([{,]\s*)([a-zA-Z0-9_]+)(\s*:)`).ReplaceAllString(sanitized, `$1"$2"$3`)

	// 第5步：最终验证
	var data interface{}
	if err := json.Unmarshal([]byte(sanitized), &data); err != nil {
		log.Printf("[ERROR] 修复后的JSON仍然无效: %v", err)

		// 如果修复后仍然无效，尝试最基本的修复：提取任何有效的JSON对象或数组
		validJsonPattern := regexp.MustCompile(`({[^{}]*})`)
		matches := validJsonPattern.FindAllString(sanitized, -1)
		if len(matches) > 0 {
			log.Printf("[INFO] 提取到 %d 个可能有效的JSON对象", len(matches))
			// 尝试每个匹配项
			for _, match := range matches {
				if json.Unmarshal([]byte(match), &data) == nil {
					log.Printf("[INFO] 找到有效的JSON对象子串")
					return match
				}
			}
		}

		// 如果还是失败，返回一个带有错误信息的JSON
		log.Printf("[WARN] 无法修复JSON，返回带有错误信息的对象")
		errorJson := fmt.Sprintf(`{"error":"无法解析返回的数据","original_text":"%s"}`,
			strings.Replace(strings.Replace(sanitized[:Min(100, len(sanitized))], `"`, `\"`, -1), "\n", "\\n", -1))
		return errorJson
	}

	// 如果经过所有处理后JSON有效，返回处理后的结果
	log.Printf("[INFO] JSON修复成功，最终长度: %d 字符", len(sanitized))
	return sanitized
}
