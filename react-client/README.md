# Aktuell React Client

A modern React TypeScript dashboard for visualizing real-time MongoDB change streams through the Aktuell server.

## 🚀 Features

- **Real-time WebSocket Connection**: Connect to Aktuell server for live data streaming
- **MongoDB Change Stream Visualization**: Watch database changes in real-time with syntax highlighting
- **Subscription Management**: Subscribe to specific databases and collections with optional filters
- **Connection Status Monitoring**: Visual indicators for connection state and health
- **Modern UI**: Dark theme with responsive design using Tailwind CSS
- **TypeScript**: Fully typed with comprehensive MongoDB change event interfaces

## 🏗️ Architecture

### Components

- **App.tsx**: Main application orchestrating all components and state management
- **ConnectionStatus**: WebSocket connection status indicator with visual feedback
- **SubscriptionManager**: Interface for managing MongoDB change stream subscriptions
- **ChangeEventsList**: Real-time display of MongoDB change events with expandable details
- **Stats**: Dashboard showing connection metrics, change counts, and subscription status

### Hooks

- **useAktuell**: Custom hook managing WebSocket connection, subscriptions, and change events

### Types

- **aktuell.ts**: Comprehensive TypeScript definitions for MongoDB change events and API interfaces

## 🛠️ Getting Started

### Prerequisites

- Node.js (v18 or higher)
- Aktuell Go server running on `ws://localhost:8080/ws` (default)

### Installation

```bash
npm install
```

### Development

Start the development server:

```bash
npm run dev
```

The application will be available at `http://localhost:5173`

### Build

Build for production:

```bash
npm run build
```

## 🎯 Usage

1. **Start Aktuell Server**: Ensure your Aktuell Go server is running
2. **Configure Connection**: Enter the WebSocket URL (default: `ws://localhost:8080/ws`)
3. **Connect**: Click the "Connect" button to establish WebSocket connection
4. **Add Subscriptions**: Subscribe to specific database collections to watch for changes
5. **Monitor Changes**: View real-time MongoDB change events as they occur

## 🔧 Configuration

### Server URL
The default server URL is `ws://localhost:8080/ws`. You can modify this in the connection input field.

### Subscription Filters
Add optional JSON filters when subscribing to collections:
```json
{"status": "active", "priority": {"$gte": 5}}
```

## 📁 Project Structure

```
src/
├── components/           # React components
│   ├── ConnectionStatus.tsx
│   ├── SubscriptionManager.tsx
│   ├── ChangeEventsList.tsx
│   └── Stats.tsx
├── hooks/               # Custom React hooks
│   └── useAktuell.tsx
├── types/               # TypeScript definitions
│   └── aktuell.ts
├── App.tsx             # Main application component
└── main.tsx            # Application entry point
```

## 🎨 UI Features

- **Dark Theme**: Professional dark theme optimized for long monitoring sessions
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Real-time Updates**: Live connection status and change event streaming
- **Syntax Highlighting**: JSON syntax highlighting for MongoDB documents
- **Expandable Views**: Click to expand change events for detailed inspection

## 🔗 Integration

This React client is designed to work with the Aktuell Go server. The server provides:
- WebSocket endpoint for real-time communication
- MongoDB change stream processing
- Client subscription management
- Real-time change event broadcasting

## 📝 License

MIT License - see LICENSE file for details
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```
