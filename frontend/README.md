# Crypto Analysis System - Frontend

React frontend application for the Crypto Analysis System built with Vite, TypeScript, TailwindCSS, and Lightweight Charts.

## Features

- **Real-time Chart**: Displays cryptocurrency candlestick charts with real-time updates via WebSocket
- **News Sidebar**: Shows crypto news with sentiment analysis (color-coded: green for positive, red for negative)
- **News Alignment**: Click on news items to see markers on the chart at the published time
- **Authentication**: Login page with JWT token management and automatic refresh
- **Responsive Design**: Modern UI built with TailwindCSS

## Tech Stack

- **React 19** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **TailwindCSS** - Utility-first CSS framework
- **Lightweight Charts** - TradingView's lightweight charting library
- **React Router** - Client-side routing
- **Axios** - HTTP client
- **js-cookie** - Cookie management

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- API Gateway running on `http://localhost:8080`

### Installation

1. Install dependencies:
```bash
npm install
```

2. Create `.env` file (copy from `.env.example`):
```bash
cp .env.example .env
```

3. Update `.env` with your API Gateway URL if different:
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080
```

### Development

Start the development server:
```bash
npm run dev
```

The app will be available at `http://localhost:3000`

### Build

Build for production:
```bash
npm run build
```

Preview production build:
```bash
npm run preview
```

## Project Structure

```
src/
├── api/
│   └── client.ts          # API client with JWT refresh logic
├── components/
│   ├── Chart.tsx          # Chart component with WebSocket
│   ├── Dashboard.tsx      # Main dashboard layout
│   ├── Login.tsx          # Login page
│   ├── NewsSidebar.tsx    # News sidebar component
│   └── ProtectedRoute.tsx # Route protection
├── types/
│   └── index.ts           # TypeScript type definitions
├── config.ts              # Configuration constants
├── App.tsx                # Main app component with routing
└── main.tsx               # Entry point
```

## API Endpoints

The frontend communicates with the API Gateway:

- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - User logout
- `GET /market/api/v1/market/history` - Get historical market data
- `GET /news/articles` - Get news articles
- `WS /ws/prices` - WebSocket for real-time price updates

## Features Details

### Chart Component

- Fetches historical candlestick data from `/market/api/v1/market/history`
- Connects to WebSocket at `/ws/prices` for real-time updates
- Updates the last candle in real-time
- Supports multiple symbols (BTCUSDT, ETHUSDT, BNBUSDT) and intervals (1m, 5m, 15m, 1h, 4h, 1d)
- Shows markers when news items are clicked

### News Sidebar

- Fetches news from `/news/articles`
- Displays title, source, time, and sentiment
- Color codes sentiment: green (positive), red (negative), gray (neutral)
- Clicking a news item adds a marker on the chart at the published time
- Auto-refreshes every 5 minutes

### Authentication

- Login form calls `/auth/login`
- JWT tokens stored in cookies (backend sets httpOnly cookies)
- Automatic token refresh on 401 errors
- Protected routes redirect to login if not authenticated

## Environment Variables

- `VITE_API_BASE_URL` - Base URL for REST API (default: http://localhost:8080)
- `VITE_WS_BASE_URL` - Base URL for WebSocket (default: ws://localhost:8080)

## License

MIT
