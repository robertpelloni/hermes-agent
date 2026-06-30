package hermes;

import java.io.InputStream;
import java.io.OutputStream;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.Arrays;
import java.util.List;
import java.util.ArrayList;

class McpPlugin {
    String name;
    Process process;
    OutputStream stdin;
    InputStream stdout;
}

public class Mcp {
    private final Map<String, McpPlugin> plugins = new ConcurrentHashMap<>();

    public void load(String name, String command, String[] args) throws Exception {
        List<String> commandList = new ArrayList<>();
        commandList.add(command);
        commandList.addAll(Arrays.asList(args));

        ProcessBuilder builder = new ProcessBuilder(commandList);
        Process process = builder.start();

        McpPlugin plugin = new McpPlugin();
        plugin.name = name;
        plugin.process = process;
        plugin.stdin = process.getOutputStream();
        plugin.stdout = process.getInputStream();

        plugins.put(name, plugin);
        System.out.println("[hermes:mcp] Loaded plugin: " + name);
    }

    public void unload(String name) {
        McpPlugin plugin = plugins.remove(name);
        if (plugin != null) {
            plugin.process.destroyForcibly();
            System.out.println("[hermes:mcp] Unloaded plugin: " + name);
        }
    }
}
