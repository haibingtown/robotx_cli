#!/bin/bash

# RobotX CLI Demo Script
# This script demonstrates the basic usage of RobotX CLI

set -e

echo "🚀 RobotX CLI Demo"
echo "=================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if robotx is installed
if ! command -v robotx &> /dev/null; then
    echo -e "${YELLOW}⚠️  robotx command not found. Building...${NC}"
    make build
    ROBOTX_CMD="./robotx"
else
    ROBOTX_CMD="robotx"
fi

echo -e "${GREEN}✓${NC} RobotX CLI is ready"
echo ""

# Show version
echo -e "${BLUE}1. Checking version${NC}"
$ROBOTX_CMD --version
echo ""

# Show help
echo -e "${BLUE}2. Available commands${NC}"
$ROBOTX_CMD --help
echo ""

# Create a demo project
echo -e "${BLUE}3. Creating demo project${NC}"
DEMO_DIR="/tmp/robotx-demo-$(date +%s)"
mkdir -p "$DEMO_DIR"

cat > "$DEMO_DIR/package.json" << 'EOF'
{
  "name": "robotx-demo",
  "version": "1.0.0",
  "description": "Demo application for RobotX CLI",
  "main": "index.js",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {
    "express": "^4.18.2"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}
EOF

cat > "$DEMO_DIR/index.js" << 'EOF'
const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({
    message: 'Hello from RobotX Demo!',
    timestamp: new Date().toISOString(),
    version: '1.0.0',
    environment: process.env.NODE_ENV || 'development'
  });
});

app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    uptime: process.uptime()
  });
});

app.listen(port, () => {
  console.log(`🚀 Server running on port ${port}`);
});
EOF

cat > "$DEMO_DIR/Dockerfile" << 'EOF'
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install --production

COPY . .

EXPOSE 3000

CMD ["npm", "start"]
EOF

echo -e "${GREEN}✓${NC} Demo project created at: $DEMO_DIR"
echo ""

# Show project structure
echo -e "${BLUE}4. Project structure${NC}"
tree "$DEMO_DIR" 2>/dev/null || ls -la "$DEMO_DIR"
echo ""

# Show deploy command help
echo -e "${BLUE}5. Deploy command options${NC}"
$ROBOTX_CMD deploy --help
echo ""

# Example: Deploy command (dry-run)
echo -e "${BLUE}6. Example deploy command${NC}"
echo -e "${YELLOW}Note: This is a dry-run example. To actually deploy, configure your RobotX server first.${NC}"
echo ""
echo "Command:"
echo "  $ROBOTX_CMD deploy $DEMO_DIR --name robotx-demo --publish"
echo ""

# Show status command help
echo -e "${BLUE}7. Status command options${NC}"
$ROBOTX_CMD status --help
echo ""

# Show logs command help
echo -e "${BLUE}8. Logs command options${NC}"
$ROBOTX_CMD logs --help
echo ""

# Show configuration example
echo -e "${BLUE}9. Configuration example${NC}"
echo "Create ~/.robotx.yaml with:"
echo ""
cat << 'EOF'
base_url: https://your-robotx-server.com
api_key: your-api-key-here
default_visibility: private
default_timeout: 600
EOF
echo ""

# Show environment variables
echo -e "${BLUE}10. Environment variables${NC}"
echo "You can also use environment variables:"
echo ""
echo "  export ROBOTX_BASE_URL=https://your-robotx-server.com"
echo "  export ROBOTX_API_KEY=your-api-key-here"
echo ""

# Show MCP integration
echo -e "${BLUE}11. MCP Integration (Claude Desktop)${NC}"
echo "Add to ~/Library/Application Support/Claude/claude_desktop_config.json:"
echo ""
cat << 'EOF'
{
  "mcpServers": {
    "robotx": {
      "command": "/usr/local/bin/robotx",
      "args": ["mcp"],
      "env": {
        "ROBOTX_BASE_URL": "https://your-robotx-server.com",
        "ROBOTX_API_KEY": "your-api-key-here"
      }
    }
  }
}
EOF
echo ""

# Show Python integration example
echo -e "${BLUE}12. Python Integration Example${NC}"
cat << 'EOF'
from robotx_client import RobotXClient

# Initialize client
client = RobotXClient(
    base_url='https://your-robotx-server.com',
    api_key='your-api-key'
)

# Deploy project
result = client.deploy(
    project_path='./my-app',
    name='my-app',
    publish=True
)

print(f"Deployed to: {result['url']}")
EOF
echo ""

# Summary
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Demo Complete!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Configure your RobotX server URL and API key"
echo "  2. Try deploying the demo project: $DEMO_DIR"
echo "  3. Check out the documentation:"
echo "     - README.md: Complete documentation"
echo "     - QUICKSTART.md: 5-minute quick start"
echo "     - docs/AI_AGENT_INTEGRATION.md: AI agent integration guide"
echo ""
echo "Demo project location: $DEMO_DIR"
echo ""
echo "To deploy the demo project:"
echo "  $ROBOTX_CMD deploy $DEMO_DIR --name robotx-demo --publish"
echo ""
