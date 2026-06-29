package hermes;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class Vfs {
    private final Map<String, byte[]> files = new ConcurrentHashMap<>();

    public void writeFile(String path, byte[] content) {
        files.put(path, content);
    }

    public byte[] readFile(String path) throws IOException {
        byte[] content = files.get(path);
        if (content != null) {
            return content;
        }
        return Files.readAllBytes(Paths.get(path));
    }

    public void flush() throws IOException {
        for (Map.Entry<String, byte[]> entry : files.entrySet()) {
            Path path = Paths.get(entry.getKey());
            Files.write(path, entry.getValue());
            System.out.println("[hermes:vfs] Flushed: " + entry.getKey());
        }
        files.clear();
    }
}
