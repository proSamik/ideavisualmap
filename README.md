# Visual Brainstorming Platform

A cloud-based SaaS platform that enables beginner solopreneurs to create visual mind maps for brainstorming ideas in targeted domains. Built with Next.js and Go, featuring AI-powered idea generation, drag-and-drop mind mapping, and a scalable architecture.

## Project Structure

This project is organized into three main directories:

- `client/`: Next.js frontend application
- `server/`: Go backend API
- `admin-client/`: Next.js admin dashboard

## Features

- 🧠 AI-powered idea generation using OpenAI API
- 🗺️ Visual mind maps with drag-and-drop interface
- 🔑 Bring your own API key option for AI integration
- 📤 Export options (PNG, PDF, JSON)
- 🔐 Secure authentication with email/password and OAuth (Google, GitHub)
- 🎨 Light/dark theme support
- 🚀 Modern, responsive UI built with Tailwind CSS
- 🔄 Real-time updates with toast notifications
- 📱 Mobile-friendly design
- 🛡️ Protected routes and API endpoints
- 🔑 Password recovery functionality
- 🔄 Session management
- 👤 User profile management
- 💰 Subscription management with LemonSqueezy

## Prerequisites

- Node.js (v18 or later)
- Go (v1.19 or later)
- PostgreSQL

## Getting Started

1. Clone the repository:
```bash
git clone <repository-url>
cd saas
```

2. Set up the frontend (client):
```bash
cd client
npm install
npm run dev
```

3. Set up the backend (server):
```bash
cd server
cp .env.example .env  # Configure your environment variables
go mod download
go run main.go
```

## Development

Refer to the README files in the `client/` and `server/` directories for detailed development guidelines and setup instructions.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
