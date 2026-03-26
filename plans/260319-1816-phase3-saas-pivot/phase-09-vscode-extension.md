---
phase: "3.9"
title: "VSCode Extension"
priority: P2
status: pending
effort: 8h
---

# Phase 3.9: VSCode Extension

## Context Links
- [MCP server client](../../mcp-server/src/client.rs) — API client pattern for Valt backend
- [CLI config](../../cli/) — agent token auth pattern
- [System architecture](../../docs/system-architecture.md) — API route reference

## Overview

VSCode extension providing TreeView sidebar, status bar notifications, and terminal integration for Valt secrets. Uses agent token (stored in VSCode SecretStorage backed by OS keychain) to communicate with Valt Go API. TypeScript strict mode, bundled with esbuild.

## Key Insights
- VSCode SecretStorage wraps OS keychain (same security as CLI keychain)
- TreeView API for sidebar: projects → secrets hierarchy
- StatusBarItem for pending approval count
- Terminal integration: inject env vars before command execution
- Extension can reuse same Go API endpoints as MCP server and CLI
- No new backend work needed — all APIs exist

## Requirements

### Functional
- **Auth**: store/retrieve agent token in VSCode SecretStorage
- **TreeView sidebar**: projects → secrets tree, expandable
- **Status bar**: show count of pending approval requests
- **Context menu**: right-click secret → "Request Access" → opens browser to dashboard
- **Command palette**: "Valt: Configure Token", "Valt: Refresh", "Valt: Inject Env"
- **Terminal integration**: `valt-inject-env` command prefix that injects approved credentials as env vars

### Non-Functional
- TypeScript strict mode
- Bundled with esbuild (single file output)
- Extension size < 500KB
- API calls with 10s timeout
- Degrade gracefully when backend unreachable

## Architecture

```
VSCode Extension (TypeScript)
  ├─ extension.ts          — activation, command registration
  ├─ auth.ts               — SecretStorage token management
  ├─ api-client.ts         — HTTP client to Valt Go API
  ├─ tree-provider.ts      — TreeDataProvider for sidebar
  ├─ status-bar.ts         — pending count polling
  ├─ terminal.ts           — env injection terminal profile
  └─ package.json          — contributes: views, commands, menus
        │
        ▼ HTTPS (Bearer agent token)
  Valt Go API (/api/v1/...)
```

## Related Code Files

### Create (all new — new directory)
- `vscode-extension/package.json` — extension manifest
- `vscode-extension/tsconfig.json`
- `vscode-extension/esbuild.mjs` — build script
- `vscode-extension/src/extension.ts` — entry point
- `vscode-extension/src/auth.ts` — token storage
- `vscode-extension/src/api-client.ts` — Valt API client
- `vscode-extension/src/tree-provider.ts` — TreeView data provider
- `vscode-extension/src/status-bar.ts` — status bar item
- `vscode-extension/src/terminal.ts` — env injection
- `vscode-extension/src/types.ts` — API response types
- `vscode-extension/.vscodeignore`
- `vscode-extension/README.md`

### No backend modifications needed

## Implementation Steps

### 1. Project scaffold

```
vscode-extension/
├── .vscodeignore
├── package.json
├── tsconfig.json
├── esbuild.mjs
└── src/
    ├── extension.ts
    ├── auth.ts
    ├── api-client.ts
    ├── tree-provider.ts
    ├── status-bar.ts
    ├── terminal.ts
    └── types.ts
```

### 2. package.json

```json
{
  "name": "valt-secrets",
  "displayName": "Valt Secrets",
  "description": "Access Valt secret vault from VSCode",
  "version": "0.1.0",
  "publisher": "valt-dev",
  "engines": { "vscode": "^1.85.0" },
  "categories": ["Other"],
  "activationEvents": ["onStartupFinished"],
  "main": "./dist/extension.js",
  "contributes": {
    "viewsContainers": {
      "activitybar": [{
        "id": "valt",
        "title": "Valt Secrets",
        "icon": "resources/valt-icon.svg"
      }]
    },
    "views": {
      "valt": [{
        "id": "valtSecrets",
        "name": "Secrets"
      }]
    },
    "commands": [
      { "command": "valt.configureToken", "title": "Valt: Configure Token" },
      { "command": "valt.refresh", "title": "Valt: Refresh" },
      { "command": "valt.requestAccess", "title": "Valt: Request Access" },
      { "command": "valt.injectEnv", "title": "Valt: Inject Env to Terminal" }
    ],
    "menus": {
      "view/item/context": [
        { "command": "valt.requestAccess", "when": "viewItem == secret", "group": "inline" }
      ]
    },
    "configuration": {
      "title": "Valt",
      "properties": {
        "valt.serverUrl": {
          "type": "string",
          "default": "http://localhost:8080",
          "description": "Valt API server URL"
        },
        "valt.pollInterval": {
          "type": "number",
          "default": 60,
          "description": "Status bar poll interval in seconds"
        }
      }
    }
  }
}
```

### 3. src/auth.ts (~40 lines)

```typescript
import * as vscode from 'vscode';

const TOKEN_KEY = 'valt-agent-token';

export class Auth {
  constructor(private secrets: vscode.SecretStorage) {}

  async getToken(): Promise<string | undefined> {
    return this.secrets.get(TOKEN_KEY);
  }

  async setToken(token: string): Promise<void> {
    await this.secrets.store(TOKEN_KEY, token);
  }

  async clearToken(): Promise<void> {
    await this.secrets.delete(TOKEN_KEY);
  }

  async requireToken(): Promise<string> {
    const token = await this.getToken();
    if (!token) {
      const input = await vscode.window.showInputBox({
        prompt: 'Enter your Valt agent token',
        password: true,
        ignoreFocusOut: true,
      });
      if (!input) throw new Error('Token required');
      await this.setToken(input);
      return input;
    }
    return token;
  }
}
```

### 4. src/api-client.ts (~80 lines)

```typescript
import { Auth } from './auth';
import type { Project, Secret, AccessRequest } from './types';

export class ValtClient {
  constructor(private baseUrl: string, private auth: Auth) {}

  private async fetch<T>(path: string, init?: RequestInit): Promise<T> {
    const token = await this.auth.requireToken();
    const res = await fetch(`${this.baseUrl}/api/v1${path}`, {
      ...init,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
        ...(init?.headers || {}),
      },
      signal: AbortSignal.timeout(10000),
    });
    if (!res.ok) throw new Error(`API error: ${res.status}`);
    return res.json();
  }

  async listProjects(): Promise<Project[]> { /* GET /projects */ }
  async listSecrets(projectId: string): Promise<Secret[]> { /* GET /secrets?project_id=X */ }
  async listPendingRequests(): Promise<{ requests: AccessRequest[]; total: number }> {
    return this.fetch('/access-requests?status=pending');
  }
  async getActiveCredentials(projectId: string): Promise<{ credentials: Credential[] }> {
    return this.fetch(`/credentials/active?project_id=${projectId}`);
  }
}
```

### 5. src/tree-provider.ts (~90 lines)

```typescript
export class ValtTreeProvider implements vscode.TreeDataProvider<ValtTreeItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<void>();
  onDidChangeTreeData = this._onDidChangeTreeData.event;

  constructor(private client: ValtClient) {}

  refresh() { this._onDidChangeTreeData.fire(); }

  getTreeItem(element: ValtTreeItem): vscode.TreeItem { return element; }

  async getChildren(element?: ValtTreeItem): Promise<ValtTreeItem[]> {
    if (!element) {
      // Root: list projects
      const projects = await this.client.listProjects();
      return projects.map(p => new ValtTreeItem(p.name, p.id, 'project',
        vscode.TreeItemCollapsibleState.Collapsed));
    }
    if (element.contextValue === 'project') {
      // Project children: list secrets
      const secrets = await this.client.listSecrets(element.resourceId);
      return secrets.map(s => new ValtTreeItem(s.name, s.id, 'secret',
        vscode.TreeItemCollapsibleState.None));
    }
    return [];
  }
}

class ValtTreeItem extends vscode.TreeItem {
  constructor(label: string, public resourceId: string, public contextValue: string,
    collapsibleState: vscode.TreeItemCollapsibleState) {
    super(label, collapsibleState);
    this.tooltip = `${contextValue}: ${label}`;
    this.iconPath = contextValue === 'project'
      ? new vscode.ThemeIcon('folder')
      : new vscode.ThemeIcon('key');
  }
}
```

### 6. src/status-bar.ts (~40 lines)

```typescript
export class StatusBar {
  private item: vscode.StatusBarItem;
  private timer: NodeJS.Timeout | undefined;

  constructor(private client: ValtClient, private intervalSec: number) {
    this.item = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    this.item.command = 'valt.refresh';
    this.item.show();
  }

  start() {
    this.poll();
    this.timer = setInterval(() => this.poll(), this.intervalSec * 1000);
  }

  stop() { if (this.timer) clearInterval(this.timer); }

  private async poll() {
    try {
      const { total } = await this.client.listPendingRequests();
      this.item.text = total > 0 ? `$(shield) Valt: ${total} pending` : '$(shield) Valt';
      this.item.tooltip = `${total} pending approval requests`;
    } catch {
      this.item.text = '$(shield) Valt: offline';
    }
  }
}
```

### 7. src/terminal.ts (~50 lines)

```typescript
export async function injectEnvToTerminal(client: ValtClient) {
  // Fetch all active credentials across projects
  const projects = await client.listProjects();
  const envVars: Record<string, string> = {};

  for (const project of projects) {
    const { credentials } = await client.getActiveCredentials(project.id);
    for (const cred of credentials) {
      // Convert secret name to ENV_VAR format: "prod-db-password" → "PROD_DB_PASSWORD"
      const envKey = cred.secret_name.toUpperCase().replace(/[^A-Z0-9]/g, '_');
      envVars[envKey] = cred.value;
    }
  }

  if (Object.keys(envVars).length === 0) {
    vscode.window.showInformationMessage('No active credentials to inject');
    return;
  }

  // Create terminal with env vars
  const terminal = vscode.window.createTerminal({
    name: 'Valt Env',
    env: envVars,
  });
  terminal.show();
  vscode.window.showInformationMessage(
    `Injected ${Object.keys(envVars).length} credentials into terminal`
  );
}
```

### 8. src/extension.ts (~60 lines)

```typescript
export function activate(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration('valt');
  const serverUrl = config.get<string>('serverUrl', 'http://localhost:8080');
  const pollInterval = config.get<number>('pollInterval', 60);

  const auth = new Auth(context.secrets);
  const client = new ValtClient(serverUrl, auth);
  const treeProvider = new ValtTreeProvider(client);
  const statusBar = new StatusBar(client, pollInterval);

  // Register tree view
  vscode.window.registerTreeDataProvider('valtSecrets', treeProvider);

  // Register commands
  context.subscriptions.push(
    vscode.commands.registerCommand('valt.configureToken', async () => {
      const token = await vscode.window.showInputBox({ prompt: 'Agent token', password: true });
      if (token) { await auth.setToken(token); treeProvider.refresh(); }
    }),
    vscode.commands.registerCommand('valt.refresh', () => treeProvider.refresh()),
    vscode.commands.registerCommand('valt.requestAccess', (item: ValtTreeItem) => {
      // Open browser to dashboard approval page
      vscode.env.openExternal(vscode.Uri.parse(`${serverUrl.replace(':8080', ':3000')}/approvals`));
    }),
    vscode.commands.registerCommand('valt.injectEnv', () => injectEnvToTerminal(client)),
  );

  // Start polling
  statusBar.start();
  context.subscriptions.push({ dispose: () => statusBar.stop() });
}

export function deactivate() {}
```

### 9. Build setup

`tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022", "module": "commonjs", "lib": ["ES2022"],
    "strict": true, "outDir": "dist", "rootDir": "src",
    "esModuleInterop": true, "skipLibCheck": true,
    "moduleResolution": "node"
  },
  "include": ["src"]
}
```

`esbuild.mjs`:
```javascript
import * as esbuild from 'esbuild';
await esbuild.build({
  entryPoints: ['src/extension.ts'],
  bundle: true, outfile: 'dist/extension.js',
  external: ['vscode'], format: 'cjs', platform: 'node',
  target: 'node18', sourcemap: true, minify: process.argv.includes('--production'),
});
```

### 10. Package scripts

```json
"scripts": {
  "build": "node esbuild.mjs",
  "build:prod": "node esbuild.mjs --production",
  "watch": "node esbuild.mjs --watch",
  "package": "vsce package"
}
```

## Todo

- [ ] Create vscode-extension/ directory scaffold
- [ ] package.json with contributes configuration
- [ ] tsconfig.json + esbuild.mjs
- [ ] Implement auth.ts (SecretStorage)
- [ ] Implement api-client.ts (HTTP client)
- [ ] Implement tree-provider.ts (TreeView)
- [ ] Implement status-bar.ts (polling)
- [ ] Implement terminal.ts (env injection)
- [ ] Implement extension.ts (activation)
- [ ] Add valt-icon.svg for activity bar
- [ ] Build and test locally (F5 Extension Host)
- [ ] Package as .vsix

## Success Criteria
- Extension activates, shows Valt icon in activity bar
- Token stored securely in OS keychain via SecretStorage
- TreeView shows projects → secrets hierarchy
- Status bar shows pending approval count
- "Request Access" opens browser to dashboard
- "Inject Env" creates terminal with active credentials as env vars
- Extension bundles to single file < 500KB

## Security Considerations
- Agent token stored in VSCode SecretStorage (OS keychain backed)
- Token never written to disk in plaintext (no settings.json)
- API calls use HTTPS in production
- Credentials in terminal env vars are ephemeral (cleared on terminal close)
- No credential values displayed in TreeView (only names)
- 10s request timeout prevents hung connections
