# RobotX CLI 客户端库示例

这个目录包含了各种编程语言的 RobotX CLI 客户端库，方便 AI agents 和自动化脚本集成。

## 可用的客户端库

### 1. Python 客户端 (`robotx_client.py`)

完整的 Python 包装器，提供 Pythonic 接口。

**安装依赖**：
```bash
# 无需额外依赖，使用 Python 标准库
```

**基本使用**：
```python
from robotx_client import RobotXClient

# 初始化客户端
client = RobotXClient(
    base_url='https://api.robotx.xin',
    api_key='your-api-key'
)

# 部署项目
result = client.deploy(
    project_path='./my-app',
    name='my-app',
    publish=True
)

print(f"Deployed to: {result['url']}")
```

**快速使用**：
```python
from robotx_client import deploy

result = deploy('./my-app', 'my-app', publish=True)
```

**功能**：
- ✅ 完整的类型提示
- ✅ 详细的错误处理
- ✅ 异步构建等待
- ✅ 便捷函数
- ✅ 命令行支持

### 2. TypeScript/Node.js 客户端 (`robotx_client.ts`)

TypeScript 客户端，提供完整的类型定义。

**安装依赖**：
```bash
npm install --save-dev typescript @types/node
```

**基本使用**：
```typescript
import { RobotXClient } from './robotx_client';

// 初始化客户端
const client = new RobotXClient({
  baseUrl: 'https://api.robotx.xin',
  apiKey: 'your-api-key'
});

// 部署项目
const result = await client.deploy('./my-app', {
  name: 'my-app',
  publish: true
});

console.log(`Deployed to: ${result.url}`);
```

**快速使用**：
```typescript
import { deploy } from './robotx_client';

const result = await deploy('./my-app', 'my-app', { publish: true });
```

**功能**：
- ✅ 完整的 TypeScript 类型
- ✅ Promise-based API
- ✅ 详细的错误处理
- ✅ 异步构建等待
- ✅ 便捷函数

### 3. JavaScript 客户端

可以直接使用 TypeScript 客户端编译后的 JavaScript 版本：

```bash
# 编译 TypeScript
tsc robotx_client.ts

# 使用编译后的 JavaScript
node robotx_client.js ./my-app my-app
```

或者使用 CommonJS 风格：

```javascript
const { RobotXClient } = require('./robotx_client');

const client = new RobotXClient({
  baseUrl: process.env.ROBOTX_BASE_URL,
  apiKey: process.env.ROBOTX_API_KEY
});

client.deploy('./my-app', { name: 'my-app', publish: true })
  .then(result => {
    console.log(`Deployed to: ${result.url}`);
  })
  .catch(error => {
    console.error('Deployment failed:', error);
  });
```

## 完整示例

### Python 示例：AI Agent 集成

```python
#!/usr/bin/env python3
"""
AI Agent with RobotX deployment capability
"""

import os
import logging
from robotx_client import RobotXClient, RobotXError

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class AIAgent:
    """AI Agent with deployment capability"""

    def __init__(self):
        self.robotx = RobotXClient(
            base_url=os.getenv('ROBOTX_BASE_URL'),
            api_key=os.getenv('ROBOTX_API_KEY')
        )

    def create_and_deploy_app(self, app_type: str, app_name: str):
        """Create and deploy an application"""
        try:
            # 1. Generate app code
            logger.info(f"Generating {app_type} application...")
            project_path = self._generate_app(app_type, app_name)

            # 2. Deploy to RobotX
            logger.info(f"Deploying to RobotX...")
            result = self.robotx.deploy(
                project_path=project_path,
                name=app_name,
                publish=True
            )

            logger.info(f"✅ Deployed successfully!")
            logger.info(f"🌐 URL: {result['url']}")

            return result

        except RobotXError as e:
            logger.error(f"❌ Deployment failed: {e}")
            raise

    def _generate_app(self, app_type: str, app_name: str) -> str:
        """Generate application code"""
        # Your code generation logic here
        pass


if __name__ == '__main__':
    agent = AIAgent()
    agent.create_and_deploy_app('web', 'my-awesome-app')
```

### TypeScript 示例：自动化部署脚本

```typescript
#!/usr/bin/env ts-node
/**
 * Automated deployment script
 */

import { RobotXClient, RobotXError } from './robotx_client';
import * as fs from 'fs';
import * as path from 'path';

interface AppConfig {
  name: string;
  type: 'web' | 'api' | 'worker';
  framework: string;
}

class DeploymentAutomation {
  private client: RobotXClient;

  constructor() {
    this.client = new RobotXClient({
      baseUrl: process.env.ROBOTX_BASE_URL,
      apiKey: process.env.ROBOTX_API_KEY
    });
  }

  async deployApp(config: AppConfig, projectPath: string): Promise<void> {
    try {
      console.log(`🚀 Deploying ${config.name}...`);

      // Deploy
      const result = await this.client.deploy(projectPath, {
        name: config.name,
        publish: false,  // Don't publish yet
        wait: true
      });

      console.log(`✅ Build completed: ${result.build_id}`);

      // Run tests (example)
      await this.runTests(result.url);

      // Publish to production
      console.log('📦 Publishing to production...');
      await this.client.publish(result.build_id);

      console.log(`🎉 Deployed successfully to: ${result.url}`);

    } catch (error) {
      if (error instanceof RobotXError) {
        console.error(`❌ Deployment failed: ${error.message}`);
        if (error.details) {
          console.error(`Details: ${error.details}`);
        }
      } else {
        console.error('❌ Unexpected error:', error);
      }
      throw error;
    }
  }

  private async runTests(url: string): Promise<void> {
    // Your test logic here
    console.log(`🧪 Running tests against ${url}...`);
  }
}

// Usage
const automation = new DeploymentAutomation();
automation.deployApp(
  {
    name: 'my-app',
    type: 'web',
    framework: 'express'
  },
  './my-app'
);
```

### Python 示例：批量部署

```python
#!/usr/bin/env python3
"""
Batch deployment script
"""

from robotx_client import RobotXClient
from concurrent.futures import ThreadPoolExecutor, as_completed
import os

def deploy_project(client: RobotXClient, project_info: dict):
    """Deploy a single project"""
    try:
        print(f"Deploying {project_info['name']}...")

        result = client.deploy(
            project_path=project_info['path'],
            name=project_info['name'],
            publish=True
        )

        return {
            'name': project_info['name'],
            'success': True,
            'url': result['url']
        }
    except Exception as e:
        return {
            'name': project_info['name'],
            'success': False,
            'error': str(e)
        }

def batch_deploy(projects: list):
    """Deploy multiple projects in parallel"""
    client = RobotXClient()

    with ThreadPoolExecutor(max_workers=3) as executor:
        futures = {
            executor.submit(deploy_project, client, proj): proj
            for proj in projects
        }

        results = []
        for future in as_completed(futures):
            result = future.result()
            results.append(result)

            if result['success']:
                print(f"✅ {result['name']}: {result['url']}")
            else:
                print(f"❌ {result['name']}: {result['error']}")

        return results

if __name__ == '__main__':
    projects = [
        {'name': 'app1', 'path': './app1'},
        {'name': 'app2', 'path': './app2'},
        {'name': 'app3', 'path': './app3'},
    ]

    results = batch_deploy(projects)

    success_count = sum(1 for r in results if r['success'])
    print(f"\n📊 Deployed {success_count}/{len(projects)} projects successfully")
```

## 错误处理

所有客户端库都提供了详细的错误处理：

### Python

```python
from robotx_client import (
    RobotXClient,
    RobotXError,
    RobotXDeploymentError,
    RobotXAPIError
)

try:
    client = RobotXClient()
    result = client.deploy('./my-app', name='my-app')
except RobotXDeploymentError as e:
    # 部署失败（构建错误等）
    print(f"Build failed: {e}")
except RobotXAPIError as e:
    # API 调用失败（认证、网络等）
    print(f"API error: {e}")
except RobotXError as e:
    # 其他错误
    print(f"Error: {e}")
```

### TypeScript

```typescript
import {
  RobotXClient,
  RobotXError,
  RobotXDeploymentError,
  RobotXAPIError
} from './robotx_client';

try {
  const client = new RobotXClient();
  const result = await client.deploy('./my-app', { name: 'my-app' });
} catch (error) {
  if (error instanceof RobotXDeploymentError) {
    // 部署失败
    console.error(`Build failed: ${error.message}`);
  } else if (error instanceof RobotXAPIError) {
    // API 调用失败
    console.error(`API error: ${error.message}`);
  } else if (error instanceof RobotXError) {
    // 其他错误
    console.error(`Error: ${error.message}`);
  }
}
```

## 高级用法

### 异步部署和状态轮询

```python
# Python
client = RobotXClient()

# 启动部署但不等待
result = client.deploy('./my-app', name='my-app', wait=False)
build_id = result['build_id']

# 做其他事情...

# 稍后等待完成
final_status = client.wait_for_build(build_id, timeout=600)
print(f"Build completed: {final_status['status']}")
```

```typescript
// TypeScript
const client = new RobotXClient();

// 启动部署但不等待
const result = await client.deploy('./my-app', {
  name: 'my-app',
  wait: false
});

// 做其他事情...

// 稍后等待完成
const finalStatus = await client.waitForBuild(result.build_id, 600);
console.log(`Build completed: ${finalStatus.status}`);
```

### 重试机制

```python
# Python with tenacity
from tenacity import retry, stop_after_attempt, wait_exponential

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10)
)
def deploy_with_retry(client, project_path, name):
    return client.deploy(project_path, name=name, publish=True)
```

```typescript
// TypeScript with retry logic
async function deployWithRetry(
  client: RobotXClient,
  projectPath: string,
  name: string,
  maxRetries: number = 3
): Promise<DeployResult> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await client.deploy(projectPath, { name, publish: true });
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, i)));
    }
  }
  throw new Error('Max retries reached');
}
```

## 测试

### Python 测试

```bash
# 运行 Python 客户端
python robotx_client.py /tmp/test-robotx-deploy test-app
```

### TypeScript 测试

```bash
# 编译并运行
tsc robotx_client.ts
node robotx_client.js /tmp/test-robotx-deploy test-app

# 或使用 ts-node
ts-node robotx_client.ts /tmp/test-robotx-deploy test-app
```

## 环境变量

所有客户端库都支持通过环境变量配置：

```bash
export ROBOTX_BASE_URL=https://api.robotx.xin
export ROBOTX_API_KEY=your-api-key-here
```

然后可以不传参数直接使用：

```python
# Python
client = RobotXClient()  # 自动从环境变量读取
```

```typescript
// TypeScript
const client = new RobotXClient();  // 自动从环境变量读取
```

## 更多资源

- [RobotX CLI 文档](../README.md)
- [快速入门](../docs/QUICKSTART.md)
- [AI Agent 集成指南](../docs/AI_AGENT_INTEGRATION.md)
- [项目总结](../docs/PROJECT_SUMMARY.md)

## 贡献

欢迎提交其他语言的客户端库实现！

## 许可证

MIT License
