import { NextResponse } from 'next/server';

// 定义后端API的URL
const BACKEND_API_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';

export async function POST(request) {
  try {
    // 获取请求中的FormData
    const formData = await request.formData();
    
    // 构建要转发到后端的FormData
    const backendFormData = new FormData();
    
    // 转发所有表单字段，包括文件和其他参数
    for (const [key, value] of formData.entries()) {
      backendFormData.append(key, value);
    }
    
    // 输出调试信息
    console.log('转发请求到后端，包含参数:', 
      Object.fromEntries(
        Array.from(formData.entries())
          .filter(([key]) => key !== 'homework')
          .map(([key, value]) => [key, value])
      )
    );
    
    // 发送请求到后端API
    const backendResponse = await fetch(`${BACKEND_API_URL}/api/homework/mark`, {
      method: 'POST',
      body: backendFormData,
    });
    
    // 获取后端响应内容
    const backendResponseData = await backendResponse.json();
    
    // 检查布局参数并处理结果
    const layout = formData.get('layout');
    
    // 如果是双栏布局且结果需要处理
    if (layout === 'double' && backendResponseData.success && 
        backendResponseData.result && backendResponseData.result.answers) {
      
      console.log('处理双栏布局结果');
      // 尝试根据位置对结果进行重新排序（模拟实现）
      // 实际实现可能需要根据后端返回的完整数据结构进行调整
      try {
        backendResponseData.result.answers = sortAnswersByColumns(backendResponseData.result.answers);
      } catch (error) {
        console.error('排序答案时出错:', error);
      }
    }
    
    // 将后端响应内容返回给前端
    return NextResponse.json(backendResponseData);
  } catch (error) {
    console.error('Error forwarding request to backend:', error);
    return NextResponse.json(
      { 
        success: false, 
        error: '请求处理失败，请稍后重试' 
      }, 
      { status: 500 }
    );
  }
}

/**
 * 将答案按照左右栏的顺序排序
 * 这是一个简化的实现，假设可以根据问题编号或位置信息排序
 */
function sortAnswersByColumns(answers) {
  if (!Array.isArray(answers) || answers.length <= 1) {
    return answers;
  }
  
  // 如果后端返回了位置信息，可以直接使用
  if (answers[0].position) {
    return answers.sort((a, b) => {
      // 首先按x坐标比较（左右栏）
      const aIsLeft = a.position.x < 0.5; // 假设0.5是页面中点
      const bIsLeft = b.position.x < 0.5;
      
      if (aIsLeft !== bIsLeft) {
        return aIsLeft ? -1 : 1; // 左栏在前
      }
      
      // 同一栏内按y坐标排序（从上到下）
      return a.position.y - b.position.y;
    });
  }
  
  // 如果没有位置信息，尝试通过题号排序
  // 假设左栏题号是连续的，右栏题号也是连续的
  const extractNumber = (questionNumber) => {
    const match = questionNumber.match(/\d+/);
    return match ? parseInt(match[0]) : 0;
  };
  
  // 提取所有题号
  const numbers = answers.map(a => extractNumber(a.questionNumber || ""));
  
  // 如果所有题号都是数字且有序
  if (numbers.every(n => n > 0)) {
    // 找出可能的分栏点
    const max = Math.max(...numbers);
    const median = Math.ceil(max / 2);
    
    // 分为左右两组
    const leftItems = answers.filter(a => extractNumber(a.questionNumber || "") <= median);
    const rightItems = answers.filter(a => extractNumber(a.questionNumber || "") > median);
    
    // 按照题号排序每组
    leftItems.sort((a, b) => extractNumber(a.questionNumber || "") - extractNumber(b.questionNumber || ""));
    rightItems.sort((a, b) => extractNumber(a.questionNumber || "") - extractNumber(b.questionNumber || ""));
    
    // 合并结果
    return [...leftItems, ...rightItems];
  }
  
  // 若无法通过题号排序，返回原始顺序
  return answers;
} 