package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Min 返回两个整数中较小的那个
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min函数类似于math.Min但适用于int类型
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// fileExists检查指定的文件或目录是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// directoryWritable检查目录是否可写
func directoryWritable(dir string) bool {
	// 创建一个临时文件来测试目录是否可写
	testFile := filepath.Join(dir, fmt.Sprintf("test_%d.tmp", time.Now().UnixNano()))
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	return true
}

// getCurrentDirectory获取当前工作目录
func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

// copyFile将一个文件复制到另一个位置
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// sortFilesByModTime按修改时间排序文件
func sortFilesByModTime(files []string) {
	sort.Slice(files, func(i, j int) bool {
		infoI, err := os.Stat(files[i])
		if err != nil {
			return false
		}

		infoJ, err := os.Stat(files[j])
		if err != nil {
			return true
		}

		return infoI.ModTime().Before(infoJ.ModTime())
	})
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

	// 首先完整清理Markdown代码块
	if strings.Contains(input, "```") {
		input = CleanMarkdownCodeBlock(input)
		log.Printf("[INFO] 已清理Markdown代码块标记")
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

	// 第3步：清理无效的UTF-8字符
	sanitized := SanitizeUTF8(input)

	// 第4步：处理括号不匹配的情况
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

	// 第5步：修复常见语法错误
	// 5.1 修复尾部多余的逗号
	sanitized = regexp.MustCompile(`,\s*[}\]]`).ReplaceAllStringFunc(sanitized, func(s string) string {
		return strings.Replace(s, ",", "", 1)
	})

	// 5.2 修复未加引号的键名
	// 这个功能需要更复杂的解析，简单起见只处理最常见的情况
	sanitized = regexp.MustCompile(`([{,]\s*)([a-zA-Z0-9_]+)(\s*:)`).ReplaceAllString(sanitized, `$1"$2"$3`)

	// 第6步：最终验证
	var data interface{}
	if err := json.Unmarshal([]byte(sanitized), &data); err != nil {
		log.Printf("[ERROR] 修复后的JSON仍然无效: %v", err)
		
		// 如果是unexpected end of JSON input错误，这可能是因为JSON被截断了
		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			log.Printf("[WARN] JSON可能被截断，重试解析完整响应")
			
			// 尝试使用更原始的处理方式
			if strings.Contains(input, "```json") && strings.Contains(input, "```") {
				// 提取```json和```之间的内容
				parts := strings.Split(input, "```json")
				if len(parts) > 1 {
					jsonContent := strings.Split(parts[1], "```")[0]
					jsonContent = strings.TrimSpace(jsonContent)
					log.Printf("[INFO] 从Markdown代码块提取JSON，长度: %d字符", len(jsonContent))
					
					// 再次处理提取出的内容
					// 验证提取的内容是否是有效的JSON
					if err := json.Unmarshal([]byte(jsonContent), &data); err == nil {
						log.Printf("[INFO] 成功从Markdown代码块提取有效JSON")
						return jsonContent
					} else {
						log.Printf("[WARN] 提取的内容仍然不是有效JSON: %v", err)
					}
				}
			}
		}

		// 如果修复后仍然无效，尝试最基本的修复：提取任何有效的JSON对象或数组
		// 尝试寻找完整的JSON对象
		objectPattern := regexp.MustCompile(`(\{[^{}]*\})`)
		objectMatches := objectPattern.FindAllString(sanitized, -1)
		
		// 也尝试寻找完整的JSON数组
		arrayPattern := regexp.MustCompile(`(\[[^\[\]]*\])`)
		arrayMatches := arrayPattern.FindAllString(sanitized, -1)
		
		allMatches := append(objectMatches, arrayMatches...)
		
		if len(allMatches) > 0 {
			log.Printf("[INFO] 提取到 %d 个可能有效的JSON片段", len(allMatches))
			// 尝试每个匹配项
			for _, match := range allMatches {
				if json.Unmarshal([]byte(match), &data) == nil {
					log.Printf("[INFO] 找到有效的JSON片段: %s", match[:Min(50, len(match))])
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

// splitPDF 分割PDF文件
func SplitPDF(inputFile string, pagesPerStudent int, outputDir string) ([]string, error) {
    log.Printf("[INFO] 开始分割PDF文件: %s, 每个学生 %d 页", inputFile, pagesPerStudent)

    // 检查输入文件是否存在
    if _, err := os.Stat(inputFile); os.IsNotExist(err) {
        log.Printf("[ERROR] 输入PDF文件不存在: %s", inputFile)
        return nil, fmt.Errorf("输入PDF文件不存在: %s", inputFile)
    }

    // 创建分割文件的目录: uploads/split
    splitDir := filepath.Join("uploads", "split")
    if err := os.MkdirAll(splitDir, 0755); err != nil {
        log.Printf("[ERROR] 创建分割文件目录失败: %s, 错误: %v", splitDir, err)
        return nil, fmt.Errorf("创建分割文件目录失败: %v", err)
    }

    log.Printf("[DEBUG] 已创建分割文件目录: %s", splitDir)

    // 创建配置
    conf := api.LoadConfiguration()

    // 获取PDF页数
    pageCount, err := api.PageCountFile(inputFile)
    if err != nil {
        log.Printf("[ERROR] 获取PDF页数失败: %v", err)
        return nil, fmt.Errorf("获取PDF页数失败: %v", err)
    }

    log.Printf("[INFO] PDF文件总页数: %d, 每个学生 %d 页", pageCount, pagesPerStudent)

    // 检查页数是否为零
    if pageCount <= 0 {
        log.Printf("[ERROR] PDF文件页数为零或无效")
        return nil, fmt.Errorf("PDF文件页数为零或无效")
    }

    // 计算需要分割的文件数
    numFiles := (pageCount + pagesPerStudent - 1) / pagesPerStudent
    var outputFiles []string

    // 获取输入文件的基本名称（不含路径和扩展名）
    baseFileName := filepath.Base(inputFile)
    fileExt := filepath.Ext(baseFileName)
    baseFileName = strings.TrimSuffix(baseFileName, fileExt)

    log.Printf("[DEBUG] 基本文件名: %s, 扩展名: %s", baseFileName, fileExt)

    // 为这次分割创建唯一的时间戳目录
    timestamp := time.Now().Format("20060102_150405")
    splitSessionDir := filepath.Join(splitDir, fmt.Sprintf("%s_%s", baseFileName, timestamp))

    if err := os.MkdirAll(splitSessionDir, 0755); err != nil {
        log.Printf("[ERROR] 创建分割会话目录失败: %s, 错误: %v", splitSessionDir, err)
        return nil, fmt.Errorf("创建分割会话目录失败: %v", err)
    }

    log.Printf("[INFO] 创建分割会话目录: %s", splitSessionDir)

    // 分割PDF - 为每个学生创建一个包含多页的PDF
    for i := 0; i < numFiles; i++ {
        startPage := i*pagesPerStudent + 1
        endPage := min((i+1)*pagesPerStudent, pageCount)

        // 创建最终输出文件名
        studentPDFName := fmt.Sprintf("student_%d.pdf", i+1)
        studentPDFFile := filepath.Join(splitSessionDir, studentPDFName)

        log.Printf("[INFO] 处理学生 %d, 页面范围: %d-%d, 输出文件: %s",
            i+1, startPage, endPage, studentPDFFile)

        // 如果之前的文件存在，先删除
        if fileExists(studentPDFFile) {
            if err := os.Remove(studentPDFFile); err != nil {
                log.Printf("[WARN] 无法删除已存在的文件: %s, 错误: %v", studentPDFFile, err)
            }
        }

        // 创建临时提取目录
        tempExtractDir := filepath.Join(splitSessionDir, fmt.Sprintf("temp_extract_%d", i+1))
        if err := os.MkdirAll(tempExtractDir, 0755); err != nil {
            return nil, fmt.Errorf("创建临时提取目录失败: %v", err)
        }

        // 构建页面范围
        var pageRanges []string
        for page := startPage; page <= endPage; page++ {
            pageRanges = append(pageRanges, fmt.Sprintf("%d", page))
        }

        log.Printf("[DEBUG] 提取页面范围: %v 到临时目录: %s", pageRanges, tempExtractDir)

        // 提取页面到临时目录 - ExtractPagesFile默认会生成带有page_*后缀的文件
        err = api.ExtractPagesFile(inputFile, tempExtractDir, pageRanges, conf)
        if err != nil {
            log.Printf("[ERROR] 提取页面 %d-%d 失败: %v", startPage, endPage, err)
            os.RemoveAll(tempExtractDir) // 清理临时目录

            // 添加临时调试日志
            log.Printf("[DEBUG] 当前工作目录: %s", getCurrentDirectory())
            log.Printf("[DEBUG] 输入文件存在: %v", fileExists(inputFile))
            log.Printf("[DEBUG] 输出目录存在: %v", fileExists(splitSessionDir))
            log.Printf("[DEBUG] 输出目录可写: %v", directoryWritable(splitSessionDir))

            // 尝试备选方法: 单独提取每一页
            log.Printf("[INFO] 尝试使用备选方法提取页面范围...")

            // 创建临时目录存放各个页面
            tempDir, err := os.MkdirTemp("", "pdf-pages-*")
            if err != nil {
                return nil, fmt.Errorf("创建临时目录失败: %v", err)
            }
            defer os.RemoveAll(tempDir)

            // 逐页提取
            var pageFiles []string
            for page := startPage; page <= endPage; page++ {
                singlePageFile := filepath.Join(tempDir, fmt.Sprintf("page_%d.pdf", page))
                log.Printf("[DEBUG] 提取单页 %d 到 %s", page, singlePageFile)

                // 提取单页
                pageRange := []string{fmt.Sprintf("%d", page)}
                if err := api.ExtractPagesFile(inputFile, singlePageFile, pageRange, conf); err != nil {
                    log.Printf("[ERROR] 提取单页 %d 失败: %v", page, err)
                    continue
                }

                // 检查生成的文件名可能有 page_ 前缀
                if !fileExists(singlePageFile) {
                    // 尝试查找真实生成的文件
                    expectedPattern := fmt.Sprintf("*page_%s.pdf", pageRange[0])
                    matches, _ := filepath.Glob(filepath.Join(tempDir, expectedPattern))
                    if len(matches) > 0 {
                        // 使用找到的文件
                        singlePageFile = matches[0]
                    } else {
                        log.Printf("[WARN] 找不到提取的页面文件，跳过")
                        continue
                    }
                }

                if fileExists(singlePageFile) {
                    pageFiles = append(pageFiles, singlePageFile)
                }
            }

            // 检查是否成功提取了页面
            if len(pageFiles) == 0 {
                return nil, fmt.Errorf("无法提取任何页面")
            }

            // 对于只有一页的情况，直接复制
            if len(pageFiles) == 1 {
                if err := copyFile(pageFiles[0], studentPDFFile); err != nil {
                    return nil, fmt.Errorf("复制单页文件失败: %v", err)
                }
            } else {
                // 对于多页，尝试使用 pdfcpu 的合并功能
                log.Printf("[INFO] 合并 %d 个页面到 %s", len(pageFiles), studentPDFFile)

                // 使用PDFCPU的合并功能 (参数：输入文件数组，输出文件，保留书签布尔值，配置)
                err = api.MergeCreateFile(pageFiles, studentPDFFile, false, conf)
                if err != nil {
                    log.Printf("[ERROR] 合并页面失败: %v", err)
                    // 如果合并失败，至少保留第一页
                    if err := copyFile(pageFiles[0], studentPDFFile); err != nil {
                        return nil, fmt.Errorf("无法保存任何页面: %v", err)
                    }
                    log.Printf("[WARN] 合并失败，仅保留第一页")
                }
            }
        } else {
            // 提取成功，处理临时目录中的文件
            log.Printf("[INFO] 页面提取成功，查找提取的文件")

            // 查找所有生成的PDF文件
            extractedFiles, _ := filepath.Glob(filepath.Join(tempExtractDir, "*.pdf"))
            if len(extractedFiles) == 0 {
                log.Printf("[ERROR] 临时目录中未找到提取的PDF文件")
                os.RemoveAll(tempExtractDir) // 清理临时目录
                return nil, fmt.Errorf("未找到提取的PDF文件")
            }

            // 按修改时间排序文件，确保页面顺序正确
            sortFilesByModTime(extractedFiles)

            // 如果只有一个文件，直接复制
            if len(extractedFiles) == 1 {
                log.Printf("[INFO] 只有一个提取文件，直接复制到 %s", studentPDFFile)
                if err := copyFile(extractedFiles[0], studentPDFFile); err != nil {
                    os.RemoveAll(tempExtractDir) // 清理临时目录
                    return nil, fmt.Errorf("复制文件失败: %v", err)
                }
            } else if len(extractedFiles) > 1 {
                // 如果有多个文件，需要合并
                log.Printf("[INFO] 合并 %d 个提取文件到 %s", len(extractedFiles), studentPDFFile)
                err = api.MergeCreateFile(extractedFiles, studentPDFFile, false, conf)
                if err != nil {
                    log.Printf("[ERROR] 合并提取文件失败: %v", err)
                    // 如果合并失败，至少保留第一个文件
                    if err := copyFile(extractedFiles[0], studentPDFFile); err != nil {
                        os.RemoveAll(tempExtractDir) // 清理临时目录
                        return nil, fmt.Errorf("无法保存任何页面: %v", err)
                    }
                    log.Printf("[WARN] 合并失败，仅保留第一个文件")
                }
            }

            // 清理临时目录
            os.RemoveAll(tempExtractDir)
        }

        // 验证输出文件存在
        if !fileExists(studentPDFFile) {
            log.Printf("[ERROR] 输出文件未创建: %s", studentPDFFile)
            return nil, fmt.Errorf("输出文件未创建: %s", studentPDFFile)
        }

        // 验证输出文件大小
        fileInfo, err := os.Stat(studentPDFFile)
        if err != nil {
            log.Printf("[WARN] 无法获取输出文件信息: %s, 错误: %v", studentPDFFile, err)
        } else if fileInfo.Size() == 0 {
            log.Printf("[ERROR] 输出文件大小为零: %s", studentPDFFile)
            return nil, fmt.Errorf("输出文件大小为零: %s", studentPDFFile)
        } else {
            log.Printf("[DEBUG] 输出文件创建成功: %s, 大小: %d 字节", studentPDFFile, fileInfo.Size())
        }

        // 确认确实提取了正确数量的页面
        extractedPageCount, err := api.PageCountFile(studentPDFFile)
        if err != nil {
            log.Printf("[WARN] 无法获取提取文件的页数: %v", err)
        } else {
            expectedPageCount := endPage - startPage + 1
            log.Printf("[INFO] 提取的文件页数: %d, 预期页数: %d", extractedPageCount, expectedPageCount)

            if extractedPageCount != expectedPageCount {
                log.Printf("[WARN] 提取的页数与预期不符！")
            }
        }

        outputFiles = append(outputFiles, studentPDFFile)
        log.Printf("[INFO] 成功创建学生 %d 的PDF文件", i+1)
    }

    log.Printf("[INFO] PDF分割完成，共创建 %d 个文件, 保存在目录: %s", len(outputFiles), splitSessionDir)
    return outputFiles, nil
}

// 拆分PDF文件为多个单独的PDF文件，每个文件包含指定页数
func SplitPDFByStudents(pdfPath string, pagesPerStudent int) ([]string, error) {
	log.Printf("[INFO] 开始按学生拆分PDF，每个学生 %d 页...", pagesPerStudent)
	
	// 检查PDF路径
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PDF文件不存在: %s", err)
	}
	
	// 获取PDF页数
	pageCount, err := api.PageCountFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("无法获取PDF页数: %s", err)
	}
	log.Printf("[INFO] PDF共有 %d 页", pageCount)
	
	// 计算学生数量
	studentCount := pageCount / pagesPerStudent
	if pageCount % pagesPerStudent != 0 {
		studentCount++
	}
	log.Printf("[INFO] 根据页数计算出 %d 个学生", studentCount)
	
	// 创建输出目录
	outputDir := filepath.Join(filepath.Dir(pdfPath), "split")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %s", err)
	}
	
	// 存储生成的PDF文件路径
	splitFiles := make([]string, 0, studentCount)
	
	// 按照每个学生的页数拆分PDF
	for i := 0; i < studentCount; i++ {
		startPage := i*pagesPerStudent + 1 // PDF页码从1开始
		endPage := (i+1) * pagesPerStudent
		
		// 确保不超过总页数
		if endPage > pageCount {
			endPage = pageCount
		}
		
		// 构建选择页面的字符串
		pageSelection := fmt.Sprintf("%d-%d", startPage, endPage)
		outputFile := filepath.Join(outputDir, fmt.Sprintf("student_%d.pdf", i+1))
		
		log.Printf("[INFO] 提取学生 %d 的页面: %s -> %s", i+1, pageSelection, outputFile)
		
		// 使用pdfcpu提取页面到新文件
		if err := api.ExtractPagesFile(pdfPath, outputFile, []string{pageSelection}, nil); err != nil {
			return nil, fmt.Errorf("提取学生 %d 页面失败: %s", i+1, err)
		}
		
		// 验证新文件
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("生成学生 %d 的PDF文件失败", i+1)
		}
		
		splitFiles = append(splitFiles, outputFile)
	}
	
	log.Printf("[INFO] PDF拆分完成，生成了 %d 个文件", len(splitFiles))
	return splitFiles, nil
}

// 直接使用PDF进行Gemini内容生成
func GenerateContentWithPDF(client *VertexAIClient, systemInstruction, pdfPath, textPrompt string) (string, error) {
	log.Printf("[INFO] 使用PDF文件生成内容: %s", pdfPath)
	
	// 增强文件存在性检查
	if pdfPath == "" {
		log.Printf("[ERROR] PDF文件路径为空")
		return "", fmt.Errorf("PDF文件路径为空")
	}
	
	// 验证PDF文件
	fileInfo, err := os.Stat(pdfPath)
	if os.IsNotExist(err) {
		log.Printf("[ERROR] PDF文件不存在: %s", pdfPath)
		return "", fmt.Errorf("PDF文件不存在: %s", pdfPath)
	}
	
	if err != nil {
		log.Printf("[ERROR] 检查PDF文件时出错: %v", err)
		return "", fmt.Errorf("检查PDF文件时出错: %v", err)
	}
	
	// 检查文件大小是否为0
	if fileInfo.Size() == 0 {
		log.Printf("[ERROR] PDF文件大小为0字节: %s", pdfPath)
		return "", fmt.Errorf("PDF文件大小为0字节: %s", pdfPath)
	}
	
	// 获取文件MIME类型
	mimeType := "application/pdf"
	
	// 验证文件是否为有效的PDF
	pageCount, pdfErr := api.PageCountFile(pdfPath)
	if pdfErr != nil {
		log.Printf("[ERROR] 无效的PDF文件: %v", pdfErr)
		return "", fmt.Errorf("无效的PDF文件: %v", pdfErr)
	}
	
	log.Printf("[INFO] PDF文件有效，页数: %d, 文件大小: %d字节", pageCount, fileInfo.Size())
	
	// 调用VertexAI处理PDF
	log.Printf("[INFO] 发送PDF文件到AI服务进行处理")
	return client.GenerateContentWithFile(systemInstruction, pdfPath, mimeType, textPrompt)
}
