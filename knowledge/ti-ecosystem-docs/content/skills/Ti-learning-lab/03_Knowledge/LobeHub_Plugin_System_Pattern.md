# LobeHub Plugin System Pattern

> **Category**: Plugin Architecture & MCP Integration
> **Source**: LobeHub Multi-Channel Plugin System
> **Verified**: true
> **Last Updated**: 2026-05-10
> **Complexity**: Advanced

---

## 🎯 Pattern Overview

LobeHub's Plugin System pattern implements the Multi-Channel Plugin (MCP) standard for extending AI agent capabilities with external tools, services, and integrations while maintaining security and performance.

---

## 🔌 MCP Architecture

### Plugin Interface Definition
```typescript
interface MCPPlugin {
  // Metadata
  name: string;
  version: string;
  description: string;
  author: string;
  license: string;
  
  // Capabilities
  capabilities: PluginCapability[];
  
  // Tools
  tools: PluginTool[];
  
  // Resources (for knowledge bases)
  resources?: PluginResource[];
  
  // Lifecycle hooks
  onInstall?: () => Promise<void>;
  onUninstall?: () => Promise<void>;
  onEnable?: () => Promise<void>;
  onDisable?: () => Promise<void>;
}

interface PluginTool {
  name: string;
  description: string;
  parameters: z.ZodSchema;
  execute: (params: any, context: ToolContext) => Promise<any>;
  timeout?: number;
  permissions?: string[];
}

interface PluginCapability {
  type: 'tool' | 'resource' | 'auth' | 'webhook';
  name: string;
  description: string;
}
```

### Plugin Manager Implementation
```typescript
class PluginManager {
  private plugins = new Map<string, MCPPlugin>();
  private tools = new Map<string, PluginTool>();
  private permissions = new PermissionManager();
  private sandbox = new PluginSandbox();

  async loadPlugin(pluginPath: string): Promise<void> {
    try {
      // Load plugin module
      const pluginModule = await import(pluginPath);
      const plugin = pluginModule.default as MCPPlugin;
      
      // Validate plugin
      await this.validatePlugin(plugin);
      
      // Check permissions
      await this.permissions.checkPluginPermissions(plugin);
      
      // Register plugin
      this.plugins.set(plugin.name, plugin);
      
      // Register tools
      for (const tool of plugin.tools) {
        this.tools.set(tool.name, tool);
      }
      
      // Call lifecycle hooks
      if (plugin.onInstall) {
        await plugin.onInstall();
      }
      
      console.log(`Plugin ${plugin.name} loaded successfully`);
    } catch (error) {
      console.error(`Failed to load plugin from ${pluginPath}:`, error);
      throw error;
    }
  }

  async executeTool(
    toolName: string, 
    params: any, 
    context: ToolContext
  ): Promise<any> {
    const tool = this.tools.get(toolName);
    
    if (!tool) {
      throw new Error(`Tool ${toolName} not found`);
    }

    // Validate parameters
    const validatedParams = tool.parameters.parse(params);
    
    // Check permissions
    await this.permissions.checkToolPermissions(tool, context);
    
    // Execute in sandbox
    return await this.sandbox.execute(tool, validatedParams, context);
  }

  private async validatePlugin(plugin: MCPPlugin): Promise<void> {
    // Check required fields
    if (!plugin.name || !plugin.version || !plugin.tools) {
      throw new Error('Plugin missing required fields');
    }

    // Validate tools
    for (const tool of plugin.tools) {
      if (!tool.name || !tool.description || !tool.execute) {
        throw new Error(`Tool ${tool.name} missing required fields`);
      }
    }

    // Check for conflicts
    if (this.plugins.has(plugin.name)) {
      throw new Error(`Plugin ${plugin.name} already exists`);
    }

    for (const tool of plugin.tools) {
      if (this.tools.has(tool.name)) {
        throw new Error(`Tool ${tool.name} already exists`);
      }
    }
  }
}
```

---

## 🛡️ Security & Sandboxing

### Plugin Sandbox Implementation
```typescript
class PluginSandbox {
  private vm: VM;
  private allowedModules: Set<string>;
  private resourceLimits: ResourceLimits;

  constructor() {
    this.vm = new VM({
      timeout: 30000,
      sandbox: {
        console: 'inherit',
        fetch: this.createSecureFetch(),
        setTimeout: 'inherit',
        clearTimeout: 'inherit'
      }
    });
    
    this.allowedModules = new Set([
      'lodash', 'moment', 'axios', 'crypto'
    ]);
    
    this.resourceLimits = {
      maxMemory: 128 * 1024 * 1024, // 128MB
      maxCpuTime: 10000, // 10 seconds
      maxNetworkRequests: 10
    };
  }

  async execute(
    tool: PluginTool, 
    params: any, 
    context: ToolContext
  ): Promise<any> {
    const executionId = generateId();
    const startTime = Date.now();
    
    try {
      // Set up resource monitoring
      const monitor = this.createResourceMonitor(executionId);
      
      // Execute tool with timeout and limits
      const result = await Promise.race([
        this.vm.run(tool.execute.toString(), {
          params,
          context: this.sanitizeContext(context),
          executionId
        }),
        this.createTimeoutPromise(tool.timeout || 30000)
      ]);

      monitor.stop();
      return result;
    } catch (error) {
      throw new PluginExecutionError(
        `Tool ${tool.name} execution failed: ${error.message}`,
        executionId
      );
    }
  }

  private createSecureFetch() {
    return async (url: string, options?: RequestInit) => {
      // Validate URL
      if (!this.isValidUrl(url)) {
        throw new Error('Invalid URL');
      }

      // Add security headers
      const secureOptions = {
        ...options,
        headers: {
          ...options?.headers,
          'User-Agent': 'LobeHub-Plugin/1.0'
        }
      };

      return fetch(url, secureOptions);
    };
  }

  private sanitizeContext(context: ToolContext): ToolContext {
    // Remove sensitive information
    return {
      ...context,
      apiKey: undefined,
      secrets: undefined,
      internalState: undefined
    };
  }
}
```

### Permission Management
```typescript
class PermissionManager {
  private permissions = new Map<string, PluginPermissions>();

  async checkPluginPermissions(plugin: MCPPlugin): Promise<void> {
    const requiredPermissions = this.extractRequiredPermissions(plugin);
    
    for (const permission of requiredPermissions) {
      if (!this.hasPermission(permission)) {
        throw new Error(`Plugin requires permission: ${permission}`);
      }
    }
  }

  async checkToolPermissions(
    tool: PluginTool, 
    context: ToolContext
  ): Promise<void> {
    if (tool.permissions) {
      for (const permission of tool.permissions) {
        if (!this.hasPermission(permission, context)) {
          throw new Error(`Tool ${tool.name} requires permission: ${permission}`);
        }
      }
    }
  }

  private extractRequiredPermissions(plugin: MCPPlugin): string[] {
    const permissions = new Set<string>();
    
    // Extract from tools
    for (const tool of plugin.tools) {
      if (tool.permissions) {
        tool.permissions.forEach(p => permissions.add(p));
      }
    }
    
    // Extract from capabilities
    for (const capability of plugin.capabilities) {
      permissions.add(`capability:${capability.type}:${capability.name}`);
    }
    
    return Array.from(permissions);
  }
}
```

---

## 🔧 Built-in Plugin Examples

### Web Search Plugin
```typescript
const webSearchPlugin: MCPPlugin = {
  name: 'web-search',
  version: '1.0.0',
  description: 'Search the web for current information',
  author: 'LobeHub Team',
  license: 'MIT',
  
  capabilities: [
    {
      type: 'tool',
      name: 'search',
      description: 'Web search capability'
    }
  ],
  
  tools: [
    {
      name: 'web_search',
      description: 'Search the web for information',
      parameters: z.object({
        query: z.string().describe('Search query'),
        limit: z.number().min(1).max(10).default(5).describe('Number of results'),
        language: z.string().optional().describe('Search language'),
        safeSearch: z.boolean().default(true).describe('Enable safe search')
      }),
      timeout: 10000,
      permissions: ['network:read', 'search:web'],
      
      async execute(params, context) {
        const searchEngine = new SearchEngine({
          apiKey: context.secrets.searchApiKey,
          provider: 'google'
        });
        
        const results = await searchEngine.search({
          query: params.query,
          limit: params.limit,
          language: params.language,
          safeSearch: params.safeSearch
        });
        
        return {
          results: results.map(r => ({
            title: r.title,
            url: r.url,
            snippet: r.snippet,
            date: r.date,
            relevanceScore: r.score
          })),
          totalResults: results.length,
          searchTime: Date.now() - context.startTime
        };
      }
    }
  ],
  
  async onInstall() {
    console.log('Web search plugin installed');
  }
};
```

### Code Execution Plugin
```typescript
const codeExecutionPlugin: MCPPlugin = {
  name: 'code-execution',
  version: '1.0.0',
  description: 'Execute code in a secure sandbox',
  author: 'LobeHub Team',
  license: 'MIT',
  
  capabilities: [
    {
      type: 'tool',
      name: 'execute',
      description: 'Code execution capability'
    }
  ],
  
  tools: [
    {
      name: 'execute_code',
      description: 'Execute code in a secure sandbox',
      parameters: z.object({
        language: z.enum(['python', 'javascript', 'bash', 'ruby']),
        code: z.string().describe('Code to execute'),
        timeout: z.number().default(5000).describe('Execution timeout in ms')
      }),
      timeout: 15000,
      permissions: ['sandbox:execute', 'resource:cpu'],
      
      async execute(params, context) {
        const sandbox = new CodeSandbox({
          language: params.language,
          timeout: params.timeout,
          memoryLimit: 64 * 1024 * 1024 // 64MB
        });
        
        try {
          const result = await sandbox.execute(params.code);
          
          return {
            output: result.stdout,
            error: result.stderr,
            exitCode: result.exitCode,
            executionTime: result.executionTime,
            memoryUsage: result.memoryUsage
          };
        } catch (error) {
          return {
            output: '',
            error: error.message,
            exitCode: 1,
            executionTime: params.timeout,
            memoryUsage: 0
          };
        }
      }
    }
  ]
};
```

### Database Plugin
```typescript
const databasePlugin: MCPPlugin = {
  name: 'database',
  version: '1.0.0',
  description: 'Connect to and query databases',
  author: 'LobeHub Team',
  license: 'MIT',
  
  capabilities: [
    {
      type: 'tool',
      name: 'query',
      description: 'Database query capability'
    },
    {
      type: 'resource',
      name: 'tables',
      description: 'Database schema resource'
    }
  ],
  
  tools: [
    {
      name: 'execute_query',
      description: 'Execute SQL query',
      parameters: z.object({
        connection: z.string().describe('Database connection name'),
        query: z.string().describe('SQL query to execute'),
        parameters: z.array(z.any()).optional().describe('Query parameters')
      }),
      timeout: 30000,
      permissions: ['database:read', 'database:write'],
      
      async execute(params, context) {
        const connection = context.connections[params.connection];
        
        if (!connection) {
          throw new Error(`Database connection ${params.connection} not found`);
        }
        
        const db = new DatabaseConnection(connection);
        
        try {
          const result = await db.query(params.query, params.parameters);
          
          return {
            rows: result.rows,
            rowCount: result.rowCount,
            executionTime: result.executionTime,
            affectedRows: result.affectedRows
          };
        } catch (error) {
          throw new Error(`Database query failed: ${error.message}`);
        }
      }
    }
  ],
  
  resources: [
    {
      name: 'database_schema',
      description: 'Database schema information',
      type: 'json',
      async load(context: ResourceContext) {
        const connection = context.connections[context.connectionName];
        const db = new DatabaseConnection(connection);
        
        const schema = await db.getSchema();
        
        return {
          tables: schema.tables,
          views: schema.views,
          relationships: schema.relationships,
          indexes: schema.indexes
        };
      }
    }
  ]
};
```

---

## 🔄 Plugin Lifecycle Management

### Plugin Registry
```typescript
class PluginRegistry {
  private registry = new Map<string, PluginMetadata>();
  private store: PluginStore;

  async registerPlugin(plugin: MCPPlugin): Promise<void> {
    const metadata: PluginMetadata = {
      name: plugin.name,
      version: plugin.version,
      description: plugin.description,
      author: plugin.author,
      license: plugin.license,
      capabilities: plugin.capabilities,
      tools: plugin.tools.map(t => ({
        name: t.name,
        description: t.description,
        parameters: t.parameters
      })),
      installedAt: new Date().toISOString(),
      status: 'active'
    };

    this.registry.set(plugin.name, metadata);
    await this.store.saveMetadata(metadata);
  }

  async getPlugin(name: string): Promise<PluginMetadata | null> {
    let metadata = this.registry.get(name);
    
    if (!metadata) {
      metadata = await this.store.loadMetadata(name);
      if (metadata) {
        this.registry.set(name, metadata);
      }
    }
    
    return metadata || null;
  }

  async listPlugins(filter?: PluginFilter): Promise<PluginMetadata[]> {
    const plugins = Array.from(this.registry.values());
    
    if (filter) {
      return plugins.filter(plugin => this.matchesFilter(plugin, filter));
    }
    
    return plugins;
  }

  private matchesFilter(plugin: PluginMetadata, filter: PluginFilter): boolean {
    if (filter.capability && !plugin.capabilities.some(c => c.type === filter.capability)) {
      return false;
    }
    
    if (filter.author && plugin.author !== filter.author) {
      return false;
    }
    
    if (filter.status && plugin.status !== filter.status) {
      return false;
    }
    
    return true;
  }
}
```

### Plugin Updates
```typescript
class PluginUpdater {
  private registry: PluginRegistry;
  private downloader: PluginDownloader;

  async updatePlugin(pluginName: string): Promise<void> {
    const currentPlugin = await this.registry.getPlugin(pluginName);
    
    if (!currentPlugin) {
      throw new Error(`Plugin ${pluginName} not found`);
    }

    // Check for updates
    const latestVersion = await this.checkForUpdates(pluginName);
    
    if (!latestVersion || latestVersion === currentPlugin.version) {
      console.log(`Plugin ${pluginName} is up to date`);
      return;
    }

    // Download new version
    const newPlugin = await this.downloader.download(pluginName, latestVersion);
    
    // Validate new version
    await this.validateUpdate(currentPlugin, newPlugin);
    
    // Install new version
    await this.installUpdate(pluginName, newPlugin);
    
    console.log(`Plugin ${pluginName} updated to version ${latestVersion}`);
  }

  private async validateUpdate(
    current: PluginMetadata, 
    update: MCPPlugin
  ): Promise<void> {
    // Check compatibility
    if (!this.isCompatible(current.version, update.version)) {
      throw new Error(`Update ${update.version} is not compatible with current version ${current.version}`);
    }

    // Validate plugin structure
    if (update.name !== current.name) {
      throw new Error('Plugin name mismatch');
    }

    // Check for breaking changes
    const breakingChanges = this.detectBreakingChanges(current, update);
    if (breakingChanges.length > 0) {
      console.warn('Breaking changes detected:', breakingChanges);
    }
  }
}
```

---

## 📊 Performance Monitoring

### Plugin Metrics
```typescript
class PluginMetrics {
  private metrics = new Map<string, PluginMetrics>();

  recordExecution(
    pluginName: string, 
    toolName: string, 
    duration: number, 
    success: boolean
  ): void {
    const key = `${pluginName}:${toolName}`;
    const existing = this.metrics.get(key) || {
      totalExecutions: 0,
      successfulExecutions: 0,
      failedExecutions: 0,
      totalDuration: 0,
      averageDuration: 0,
      lastExecution: null
    };

    existing.totalExecutions++;
    existing.totalDuration += duration;
    existing.averageDuration = existing.totalDuration / existing.totalExecutions;
    existing.lastExecution = new Date().toISOString();

    if (success) {
      existing.successfulExecutions++;
    } else {
      existing.failedExecutions++;
    }

    this.metrics.set(key, existing);
  }

  getMetrics(pluginName?: string): PluginMetricsData[] {
    const metrics: PluginMetricsData[] = [];

    for (const [key, data] of this.metrics.entries()) {
      const [plugin, tool] = key.split(':');
      
      if (!pluginName || plugin === pluginName) {
        metrics.push({
          plugin,
          tool,
          ...data,
          successRate: data.successfulExecutions / data.totalExecutions
        });
      }
    }

    return metrics;
  }
}
```

---

## 🎯 Implementation Guidelines

### When to Use Plugin System

1. **Extensible Platform**: Need to support third-party extensions
2. **Tool Integration**: Multiple external services and APIs
3. **Custom Workflows**: User-specific functionality
4. **Ecosystem Building**: Community contributions

### Security Best Practices

1. **Sandboxing**: Always execute plugins in isolated environments
2. **Permission Control**: Fine-grained permission system
3. **Resource Limits**: Prevent resource exhaustion
4. **Input Validation**: Strict parameter validation
5. **Audit Logging**: Track all plugin activities

### Performance Considerations

1. **Lazy Loading**: Load plugins only when needed
2. **Caching**: Cache plugin metadata and results
3. **Monitoring**: Track plugin performance and usage
4. **Resource Management**: Limit memory and CPU usage
5. **Timeout Handling**: Prevent hanging plugins

---

*Pattern extracted from LobeHub MCP implementation*  
*Supports 100+ community plugins*  
*Last updated: 2026-05-10*