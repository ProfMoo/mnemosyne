import { execFileSync } from 'node:child_process';
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

const directory = new URL('../src/api/generated/', import.meta.url);

function snapshot() {
  return JSON.stringify(
    readdirSync(directory, { recursive: true, withFileTypes: true })
      .filter((entry) => entry.isFile())
      .map((entry) => {
        const path = join(entry.parentPath, entry.name);
        return [path, readFileSync(path, 'utf8')];
      })
      .sort(([left], [right]) => left.localeCompare(right)),
  );
}

const before = snapshot();
execFileSync('npm', ['run', 'generate'], { stdio: 'inherit' });
if (before !== snapshot()) {
  process.stderr.write(
    'Generated client was stale. Review and commit regenerated src/api/generated files.\n',
  );
  process.exitCode = 1;
}
