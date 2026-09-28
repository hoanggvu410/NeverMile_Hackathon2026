import { execFile } from "node:child_process";
import { promisify } from "node:util";
import path from "node:path";
import { NextResponse } from "next/server";

const execFileAsync = promisify(execFile);
const separator = "\u001f";

export const dynamic = "force-dynamic";

export async function GET() {
  const repo = process.env.GITWHY_REPO_PATH
    ? path.resolve(process.env.GITWHY_REPO_PATH)
    : path.resolve(process.cwd(), "..");

  try {
    const { stdout } = await execFileAsync(
      "git",
      [
        "-C",
        repo,
        "log",
        "--all",
        "--topo-order",
        "--date=iso-strict",
        `--format=%H${separator}%P${separator}%an${separator}%ad${separator}%s`,
        "-n",
        "150",
      ],
      { maxBuffer: 2 * 1024 * 1024 }
    );

    const commits = stdout
      .trim()
      .split(/\r?\n/)
      .filter(Boolean)
      .map((line) => {
        const [sha, parents = "", author = "", date = "", ...subject] =
          line.split(separator);
        return {
          sha,
          parents: parents ? parents.split(" ") : [],
          author,
          date,
          subject: subject.join(separator),
        };
      });

    return NextResponse.json(commits);
  } catch (error) {
    const message = error instanceof Error ? error.message : "Could not read Git history";
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
