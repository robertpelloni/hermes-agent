import * as fs from 'fs';
import * as path from 'path';

export class VirtualFileSystem {
    private files: Map<string, Buffer> = new Map();

    public writeFile(filePath: string, content: Buffer): void {
        this.files.set(filePath, content);
    }

    public async readFile(filePath: string): Promise<Buffer> {
        const content = this.files.get(filePath);
        if (content) {
            return content;
        }
        return fs.promises.readFile(filePath);
    }

    public async flush(): Promise<void> {
        for (const [filePath, content] of this.files.entries()) {
            await fs.promises.writeFile(filePath, content);
            console.log(`[hermes:vfs] Flushed: ${filePath}`);
        }
        this.files.clear();
    }
}
