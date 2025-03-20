# AI Exam Mark System

A homework marking system built with Next.js and Go, supporting handwritten text recognition and automatic scoring.

## Features

- Upload homework images
- Automatic handwritten text recognition
- Support for single-column and double-column exam layouts
- Real-time preview and feedback
- Responsive design

## Tech Stack

### Frontend
- Next.js 14
- React 18
- TypeScript
- Tailwind CSS
- shadcn/ui

### Backend
- Go
- Gin Framework
- Google Cloud Vision API
- Gemini API

## Local Development

### Prerequisites

- Node.js 18+
- Go 1.20+
- npm or yarn
- Google Cloud account and credentials

### Installation Steps

1. Clone the repository
```bash
git clone https://github.com/yourusername/homework-marking.git
cd homework-marking
```

2. Install frontend dependencies
```bash
npm install
```

3. Install backend dependencies
```bash
cd backend
go mod download
cd ..
```

4. Configure environment variables
```bash
# Copy environment variable example files
cp .env.example .env.local
cp backend/.env.example backend/.env

# Edit the environment variable files with your configuration
```

5. Configure Google Cloud credentials
- Place your Google Cloud service account key file in the `backend` directory
- Update the `GOOGLE_APPLICATION_CREDENTIALS` path in `backend/.env`

6. Start development servers
```bash
# Start backend service
cd backend
./run_server.sh

# In a new terminal, start frontend service
npm run dev
```

Visit http://localhost:3000 to view the application

## Deployment

### Frontend Deployment (Vercel)

1. Create a new project on Vercel
2. Import the GitHub repository
3. Configure environment variables
4. Deploy

### Backend Deployment

1. Prepare server environment
2. Configure environment variables
3. Compile and run the backend service

## Project Structure

```
homework-marking/
├── frontend/           # Next.js frontend application
│   ├── src/           # Source code
│   ├── public/        # Static assets
│   └── package.json   # Frontend dependencies
├── backend/           # Go backend service
│   ├── cmd/          # Main program entry
│   ├── handlers/     # Request handlers
│   ├── models/       # Data models
│   └── services/     # Business logic
└── README.md         # Project documentation
```

## Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

MIT License - See LICENSE file for details 