import { type FileOptions, file } from "tmp";
import fsPromises, { type FileHandle } from "node:fs/promises";

export async function tmpFileAsync(
  options: FileOptions,
): Promise<{ name: string; handle: FileHandle } & AsyncDisposable> {
  const { name } = await new Promise<{
    name: string;
  }>((resolve, reject) => {
    file(
      {
        ...options,
        // We close and clean the file up ourselves
        discardDescriptor: true,
        detachDescriptor: true,
        keep: true,
      },
      (err, name) => {
        if (err) {
          reject(err);
        } else {
          resolve({ name });
        }
      },
    );
  });

  // `tmp` returns a fd but it's much easier to work with a `FileHandle` so we
  // create one ourselves.
  const handle = await fsPromises.open(name, "w", options.mode ?? 0o600);

  return {
    name,
    handle,
    async [Symbol.asyncDispose]() {
      await handle.close();

      await fsPromises.unlink(name);
    },
  };
}
