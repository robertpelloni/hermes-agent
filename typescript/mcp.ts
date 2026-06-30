import { spawn, ChildProcessWithoutNullStreams } from 'child_process';

export class McpPlugin {
    constructor(
        public name: string,
        public process: ChildProcessWithoutNullStreams
    ) {}
}

export class McpPluginManager {
    private plugins: Map<string, McpPlugin> = new Map();

    public load(name: string, command: string, args: string[]): void {
        const proc = spawn(command, args);

        const plugin = new McpPlugin(name, proc);
        this.plugins.set(name, plugin);
        console.log(`[hermes:mcp] Loaded plugin: ${name}`);
    }

    public unload(name: string): void {
        const plugin = this.plugins.get(name);
        if (plugin) {
            plugin.process.kill('SIGKILL');
            this.plugins.delete(name);
            console.log(`[hermes:mcp] Unloaded plugin: ${name}`);
        }
    }
}
