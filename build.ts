import { JSZip, readZip } from "https://deno.land/x/jszip@0.11.0/mod.ts";

import { JSZipGeneratorOptions } from "https://deno.land/x/jszip@0.11.0/types.ts";
// import { filename } from "https://denopkg.com/rsp/deno-dirname@master/mod.ts";
import { resolve } from "https://deno.land/std@0.181.0/path/mod.ts";

if (import.meta.main) {
  await Promise.all([
    addFileToZipFileOutput(
      "./temp/speedtest-go-main.zip",
      "speedtest-go-main/speedtest.exe",
      "./dist/windows-amd64/speedtest.exe",
      "./dist/windows-amd64-speedtest.zip",
    ),
    addFileToZipFileOutput(
      "./temp/speedtest-go-main.zip",
      "speedtest-go-main/speedtest",
      "./dist/linux-amd64/speedtest",
      "./dist/linux-amd64-speedtest.zip",
    ),
    addFileToZipFileOutput(
      "./temp/speedtest-go-main.zip",
      "speedtest-go-main/speedtest",
      "./dist/linux-arm64/speedtest",
      "./dist/linux-arm64-speedtest.zip",
    ),
    addFileToZipFileOutput(
      "./temp/speedtest-go-main.zip",
      "speedtest-go-main/speedtest",
      "./dist/linux-mipsle/speedtest",
      "./dist/linux-mipsle-speedtest.zip",
    ),
  ]);
}
export async function addFileToZipFileOutput(
  source: string,
  path: string,
  file: string,
  dest: string,
) {
  console.log({
    source: resolve(source),
    path,
    file: resolve(file),
    dest: resolve(dest),
  });
  const zip = await readZip(resolve(source)); //new JSZip();
  // await zip.loadAsync(
  //     await (
  //         await fetch(import.meta.resolve("./temp/speedtest-go-main.zip"))
  //     ).arrayBuffer()
  // );
  zip.addFile(
    path, // "speedtest-go-main/speedtest.exe",
    await Deno.readFile(resolve(file)),
    // new Uint8Array(
    //     await (
    //         await fetch(
    //             import.meta.resolve("./dist/windows-amd64/speedtest.exe")
    //         )
    //     ).arrayBuffer()
    // )
  );
  // const { __filename, __dirname } = __(import.meta);
  // console.log(__dirname);
  // console.log(
  const outputfile = resolve(
    dest,
    /* "./dist/windows-amd64/speedtest.zip" */
  );
  // );
  // console.log(outputfile);
  await writeZip(zip, outputfile, { compression: "DEFLATE" });

  // console.log(zip);
  // console.log(outputfile);
}
export async function writeZip(
  zip: JSZip,
  path: string,
  options?: JSZipGeneratorOptions<"uint8array"> | undefined,
) {
  const b: Uint8Array = await zip.generateAsync({
    type: "uint8array",
    ...options,
  });
  return await Deno.writeFile(path, b);
}
