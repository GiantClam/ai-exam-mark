import { NextResponse } from 'next/server';

export async function POST(request: Request) {
  try {
    const formData = await request.formData();
    const file = formData.get('file') as File;
    const pagesPerStudent = formData.get('pagesPerStudent') as string;
    const type = formData.get('type') as string;

    if (!file) {
      return NextResponse.json(
        { error: '请上传PDF文件' },
        { status: 400 }
      );
    }

    // 创建新的 FormData 对象发送到后端
    const backendFormData = new FormData();
    backendFormData.append('file', file);
    backendFormData.append('pagesPerStudent', pagesPerStudent);
    backendFormData.append('type', type);

    // 发送请求到后端
    const response = await fetch('http://localhost:8280/api/homework/pdf', {
      method: 'POST',
      body: backendFormData,
    });

    if (!response.ok) {
      const errorData = await response.json();
      return NextResponse.json(
        { error: errorData.error || '批改失败' },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error marking homework:', error);
    return NextResponse.json(
      { error: '批改失败，请重试' },
      { status: 500 }
    );
  }
} 