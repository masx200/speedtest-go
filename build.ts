import { JSZip, readZip } from "https://deno.land/x/jszip@0.11.0/mod.ts";

import { JSZipGeneratorOptions } from "https://deno.land/x/jszip@0.11.0/types.ts";
import { resolve } from "https://deno.land/std@0.181.0/path/mod.ts";
import { type JSZipFileOptions } from "https://deno.land/x/jszip@0.11.0/types.ts";
if (import.meta.main) {
  const source = resolve("./temp/speedtest-go-main.zip");
  const fileoptions = { unixPermissions: "777" } satisfies JSZipFileOptions;
  const generateoptions = {
    compression: "DEFLATE",
  } satisfies JSZipGeneratorOptions<"uint8array">;
  await Promise.all([
    addFileToZipFileOutput(
      source,
      "speedtest-go-main/speedtest.exe",
      "./dist/windows-amd64/speedtest.exe",
      "./dist/windows-amd64-speedtest.zip",
      fileoptions,
      generateoptions,
    ),
    addFileToZipFileOutput(
      source,
      "speedtest-go-main/speedtest",
      "./dist/linux-amd64/speedtest",
      "./dist/linux-amd64-speedtest.zip",
      fileoptions,
      generateoptions,
    ),
    addFileToZipFileOutput(
      source,
      "speedtest-go-main/speedtest",
      "./dist/linux-arm64/speedtest",
      "./dist/linux-arm64-speedtest.zip",
      fileoptions,
      generateoptions,
    ),
    addFileToZipFileOutput(
      source,
      "speedtest-go-main/speedtest",
      "./dist/linux-mipsle-softfloat/speedtest",
      "./dist/linux-mipsle-softfloat-speedtest.zip",
      fileoptions,
      generateoptions,
    ),
  ]);
}
export async function addFileToZipFileOutput(
  source: string,
  path: string,
  file: string,
  dest: string,
  fileoptions: JSZipFileOptions = {},
  generateoptions: JSZipGeneratorOptions<"uint8array"> = {},
) {
  generateoptions = {
    compression: "DEFLATE",
    ...generateoptions,
  };
  console.log({
    source: resolve(source),
    path,
    file: resolve(file),
    dest: resolve(dest),
    fileoptions,
    generateoptions,
  });
  const zip = await readZip(resolve(source));
  zip.addFile(path, await Deno.readFile(resolve(file)), fileoptions);
  const outputfile = resolve(dest);
  await writeZip(zip, outputfile, generateoptions);
}
export async function writeZip(
  zip: JSZip,
  path: string,
  options: JSZipGeneratorOptions<"uint8array"> = {},
) {
  const b: Uint8Array = await zip.generateAsync({
    type: "uint8array",
    ...options,
  });
  return await Deno.writeFile(path, b);
}
